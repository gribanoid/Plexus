package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/plexus/backend/internal/api"
	"github.com/plexus/backend/internal/config"
	"github.com/plexus/backend/internal/db"
	"github.com/plexus/backend/internal/search"
	"github.com/plexus/backend/internal/websocket"
	"github.com/redis/go-redis/v9"
)

var testApp *fiber.App

func TestMain(m *testing.M) {
	if os.Getenv("DATABASE_URL") == "" {
		os.Exit(0)
	}

	if err := db.RunMigrations(os.Getenv("DATABASE_URL")); err != nil {
		fmt.Fprintf(os.Stderr, "migrations: %v\n", err)
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "pool: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "redis parse: %v\n", err)
		os.Exit(1)
	}
	rdb := redis.NewClient(opt)
	defer rdb.Close()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	cfg.JWTSecret = "test-jwt-secret-at-least-32-characters!!"
	cfg.Env = "test"
	cfg.MetricsEnabled = false
	cfg.AllowRegistration = true

	testApp = api.New(api.Dependencies{
		Config:       cfg,
		Pool:         pool,
		Redis:        rdb,
		SearchClient: search.NewClient(cfg.MeilisearchURL, cfg.MeilisearchKey),
		Hub:          websocket.NewHub(rdb),
		JobClient:    asynq.NewClient(asynq.RedisClientOpt{Addr: opt.Addr}),
	})

	os.Exit(m.Run())
}

func TestAuthRegisterLoginRefresh(t *testing.T) {
	email := fmt.Sprintf("user-%s@test.local", uuid.NewString()[:8])

	regBody, _ := json.Marshal(map[string]string{
		"email":        email,
		"password":     "password123",
		"display_name": "Test User",
	})
	regResp := doRequest(t, "POST", "/api/v1/auth/register", regBody, "")
	if regResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("register: expected 201, got %d: %s", regResp.StatusCode, readBody(regResp))
	}

	var tokens tokenPair
	decodeJSON(t, regResp, &tokens)
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("expected token pair")
	}

	loginBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": "password123",
	})
	loginResp := doRequest(t, "POST", "/api/v1/auth/login", loginBody, "")
	if loginResp.StatusCode != fiber.StatusOK {
		t.Fatalf("login: expected 200, got %d", loginResp.StatusCode)
	}

	refreshBody, _ := json.Marshal(map[string]string{"refresh_token": tokens.RefreshToken})
	refreshResp := doRequest(t, "POST", "/api/v1/auth/refresh", refreshBody, "")
	if refreshResp.StatusCode != fiber.StatusOK {
		t.Fatalf("refresh: expected 200, got %d: %s", refreshResp.StatusCode, readBody(refreshResp))
	}

	var newTokens tokenPair
	decodeJSON(t, refreshResp, &newTokens)
	if newTokens.RefreshToken == tokens.RefreshToken {
		t.Fatal("refresh token should rotate")
	}

	// Old refresh token must be invalid
	oldRefreshResp := doRequest(t, "POST", "/api/v1/auth/refresh", refreshBody, "")
	if oldRefreshResp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("old refresh: expected 401, got %d", oldRefreshResp.StatusCode)
	}
}

func TestACL_DeniesNonMemberOrgAccess(t *testing.T) {
	userA := registerUser(t, "a")
	userB := registerUser(t, "b")

	// User A creates org via registration; get their org slug from /orgs
	orgsResp := doRequest(t, "GET", "/api/v1/orgs", nil, userA.AccessToken)
	var orgs struct {
		Items []struct {
			Slug string `json:"slug"`
		} `json:"items"`
	}
	decodeJSON(t, orgsResp, &orgs)
	if len(orgs.Items) == 0 {
		t.Fatal("user A should have an org")
	}
	orgSlug := orgs.Items[0].Slug

	// User B cannot access user A's org
	getOrgResp := doRequest(t, "GET", "/api/v1/orgs/"+orgSlug, nil, userB.AccessToken)
	if getOrgResp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404 for non-member, got %d", getOrgResp.StatusCode)
	}
}

func TestIssuesCRUD(t *testing.T) {
	user := registerUser(t, "issues")
	orgSlug, projectKey := firstOrgProject(t, user.AccessToken)

	// List issue types to get type_id
	typesResp := doRequest(t, "GET", fmt.Sprintf("/api/v1/orgs/%s/projects/%s/issue-types", orgSlug, projectKey), nil, user.AccessToken)
	var types struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	decodeJSON(t, typesResp, &types)
	if len(types.Items) == 0 {
		t.Fatal("expected issue types")
	}

	createBody, _ := json.Marshal(map[string]interface{}{
		"title":   "Test issue",
		"type_id": types.Items[0].ID,
	})
	createResp := doRequest(t, "POST", fmt.Sprintf("/api/v1/orgs/%s/projects/%s/issues", orgSlug, projectKey), createBody, user.AccessToken)
	if createResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("create issue: expected 201, got %d: %s", createResp.StatusCode, readBody(createResp))
	}

	var created struct {
		Number int64 `json:"number"`
	}
	decodeJSON(t, createResp, &created)

	getResp := doRequest(t, "GET", fmt.Sprintf("/api/v1/orgs/%s/projects/%s/issues/%d", orgSlug, projectKey, created.Number), nil, user.AccessToken)
	if getResp.StatusCode != fiber.StatusOK {
		t.Fatalf("get issue: expected 200, got %d", getResp.StatusCode)
	}

	patchBody, _ := json.Marshal(map[string]string{"title": "Updated title"})
	patchResp := doRequest(t, "PATCH", fmt.Sprintf("/api/v1/orgs/%s/projects/%s/issues/%d", orgSlug, projectKey, created.Number), patchBody, user.AccessToken)
	if patchResp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("patch issue: expected 204, got %d", patchResp.StatusCode)
	}

	delResp := doRequest(t, "DELETE", fmt.Sprintf("/api/v1/orgs/%s/projects/%s/issues/%d", orgSlug, projectKey, created.Number), nil, user.AccessToken)
	if delResp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("delete issue: expected 204, got %d", delResp.StatusCode)
	}
}

func TestACL_GuestCannotWrite(t *testing.T) {
	owner := registerUser(t, "owner")
	guest := registerUser(t, "guest")

	orgSlug, projectKey := firstOrgProject(t, owner.AccessToken)

	guestEmail := getUserEmail(t, guest.AccessToken)
	inviteBody, _ := json.Marshal(map[string]string{
		"email": guestEmail,
		"role":  "guest",
	})
	inviteResp := doRequest(t, "POST", "/api/v1/orgs/"+orgSlug+"/members/invite", inviteBody, owner.AccessToken)
	if inviteResp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("invite guest: expected 204, got %d: %s", inviteResp.StatusCode, readBody(inviteResp))
	}

	typesResp := doRequest(t, "GET", fmt.Sprintf("/api/v1/orgs/%s/projects/%s/issue-types", orgSlug, projectKey), nil, guest.AccessToken)
	var types struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	decodeJSON(t, typesResp, &types)
	if len(types.Items) == 0 {
		t.Fatal("expected issue types")
	}

	createBody, _ := json.Marshal(map[string]interface{}{
		"title":   "Guest issue",
		"type_id": types.Items[0].ID,
	})
	createResp := doRequest(t, "POST", fmt.Sprintf("/api/v1/orgs/%s/projects/%s/issues", orgSlug, projectKey), createBody, guest.AccessToken)
	if createResp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("guest create issue: expected 403, got %d: %s", createResp.StatusCode, readBody(createResp))
	}
}

func TestSprintStartComplete(t *testing.T) {
	user := registerUser(t, "sprint")
	orgSlug, projectKey := firstOrgProject(t, user.AccessToken)

	createBody, _ := json.Marshal(map[string]string{"name": "Sprint 1"})
	createResp := doRequest(t, "POST", fmt.Sprintf("/api/v1/orgs/%s/projects/%s/sprints", orgSlug, projectKey), createBody, user.AccessToken)
	if createResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("create sprint: expected 201, got %d: %s", createResp.StatusCode, readBody(createResp))
	}

	var created struct {
		ID string `json:"id"`
	}
	decodeJSON(t, createResp, &created)

	startResp := doRequest(t, "POST", fmt.Sprintf("/api/v1/orgs/%s/projects/%s/sprints/%s/start", orgSlug, projectKey, created.ID), nil, user.AccessToken)
	if startResp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("start sprint: expected 204, got %d: %s", startResp.StatusCode, readBody(startResp))
	}

	listResp := doRequest(t, "GET", fmt.Sprintf("/api/v1/orgs/%s/projects/%s/sprints", orgSlug, projectKey), nil, user.AccessToken)
	var sprints struct {
		Items []struct {
			ID    string `json:"id"`
			State string `json:"state"`
		} `json:"items"`
	}
	decodeJSON(t, listResp, &sprints)
	foundActive := false
	for _, s := range sprints.Items {
		if s.ID == created.ID && s.State == "active" {
			foundActive = true
			break
		}
	}
	if !foundActive {
		t.Fatal("expected sprint to be active after start")
	}

	completeResp := doRequest(t, "POST", fmt.Sprintf("/api/v1/orgs/%s/projects/%s/sprints/%s/complete", orgSlug, projectKey, created.ID), nil, user.AccessToken)
	if completeResp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("complete sprint: expected 204, got %d: %s", completeResp.StatusCode, readBody(completeResp))
	}

	listResp2 := doRequest(t, "GET", fmt.Sprintf("/api/v1/orgs/%s/projects/%s/sprints", orgSlug, projectKey), nil, user.AccessToken)
	var sprints2 struct {
		Items []struct {
			ID    string `json:"id"`
			State string `json:"state"`
		} `json:"items"`
	}
	decodeJSON(t, listResp2, &sprints2)
	foundClosed := false
	for _, s := range sprints2.Items {
		if s.ID == created.ID && s.State == "closed" {
			foundClosed = true
			break
		}
	}
	if !foundClosed {
		t.Fatal("expected sprint to be closed after complete")
	}
}

func getUserEmail(t *testing.T, token string) string {
	t.Helper()
	resp := doRequest(t, "GET", "/api/v1/me", nil, token)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("get me: expected 200, got %d", resp.StatusCode)
	}
	var user struct {
		Email string `json:"email"`
	}
	decodeJSON(t, resp, &user)
	return user.Email
}

type tokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func registerUser(t *testing.T, prefix string) tokenPair {
	t.Helper()
	email := fmt.Sprintf("%s-%s@test.local", prefix, uuid.NewString()[:8])
	body, _ := json.Marshal(map[string]string{
		"email":        email,
		"password":     "password123",
		"display_name": "Test " + prefix,
	})
	resp := doRequest(t, "POST", "/api/v1/auth/register", body, "")
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("register %s: %d %s", prefix, resp.StatusCode, readBody(resp))
	}
	var tokens tokenPair
	decodeJSON(t, resp, &tokens)
	return tokens
}

func firstOrgProject(t *testing.T, token string) (orgSlug, projectKey string) {
	t.Helper()
	orgsResp := doRequest(t, "GET", "/api/v1/orgs", nil, token)
	var orgs struct {
		Items []struct {
			Slug string `json:"slug"`
		} `json:"items"`
	}
	decodeJSON(t, orgsResp, &orgs)
	if len(orgs.Items) == 0 {
		t.Fatal("no orgs")
	}
	orgSlug = orgs.Items[0].Slug

	projectsResp := doRequest(t, "GET", "/api/v1/orgs/"+orgSlug+"/projects", nil, token)
	var projects struct {
		Items []struct {
			Key string `json:"key"`
		} `json:"items"`
	}
	decodeJSON(t, projectsResp, &projects)
	if len(projects.Items) == 0 {
		t.Fatal("no projects")
	}
	return orgSlug, projects.Items[0].Key
}

func doRequest(t *testing.T, method, path string, body []byte, bearer string) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := testApp.Test(req, int((30 * time.Second).Milliseconds()))
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func decodeJSON(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode json: %v", err)
	}
}

func readBody(resp *http.Response) string {
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return string(b)
}

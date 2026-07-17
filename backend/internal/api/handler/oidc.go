package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/plexus/backend/internal/auth"
	"github.com/plexus/backend/internal/repository"
)

const (
	oidcStateCookie   = "oidc_state"
	oidcExchangeTTL   = 2 * time.Minute
	oidcExchangePrefix = "oidc:exchange:"
)

func (h *Handler) OIDCLogin(c *fiber.Ctx) error {
	if h.OIDC == nil || !h.OIDC.Configured() {
		return fiber.NewError(fiber.StatusNotImplemented, "OIDC is not configured")
	}

	state, err := auth.NewState()
	if err != nil {
		return err
	}

	authURL, err := h.OIDC.AuthorizationURL(state)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "failed to discover OIDC provider")
	}

	c.Cookie(&fiber.Cookie{
		Name:     oidcStateCookie,
		Value:    state,
		HTTPOnly: true,
		Secure:   c.Protocol() == "https" || !strings.EqualFold(c.Hostname(), "localhost"),
		SameSite: "Lax",
		MaxAge:   600,
		Path:     "/",
	})

	return c.Redirect(authURL, fiber.StatusFound)
}

func (h *Handler) OIDCCallback(c *fiber.Ctx) error {
	if h.OIDC == nil || !h.OIDC.Configured() {
		return fiber.NewError(fiber.StatusNotImplemented, "OIDC is not configured")
	}

	if errMsg := c.Query("error"); errMsg != "" {
		return fiber.NewError(fiber.StatusUnauthorized, "OIDC authorization failed")
	}

	code := c.Query("code")
	if code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "authorization code is required")
	}

	state := c.Query("state")
	cookieState := c.Cookies(oidcStateCookie)
	if state == "" || cookieState == "" || state != cookieState {
		return fiber.NewError(fiber.StatusBadRequest, "invalid OAuth state")
	}
	c.ClearCookie(oidcStateCookie)

	tokens, err := h.OIDC.ExchangeCode(c.Context(), code)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "token exchange failed")
	}

	accessToken, _ := tokens["access_token"].(string)
	var claims map[string]any
	if accessToken != "" {
		claims, err = h.OIDC.UserInfo(c.Context(), accessToken)
		if err != nil {
			return fiber.NewError(fiber.StatusBadGateway, "userinfo request failed")
		}
	}

	email, _ := claims["email"].(string)
	if email == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email claim is required")
	}
	if verified, ok := claims["email_verified"].(bool); ok && !verified {
		return fiber.NewError(fiber.StatusForbidden, "email is not verified by identity provider")
	}
	// Some providers return email_verified as string
	if verifiedStr, ok := claims["email_verified"].(string); ok && !strings.EqualFold(verifiedStr, "true") {
		return fiber.NewError(fiber.StatusForbidden, "email is not verified by identity provider")
	}

	creds, err := h.Repo.GetUserCredentials(c.Context(), email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			creds, err = h.provisionOIDCUser(c, email, claims)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	tokenPair, err := h.issueTokens(c.Context(), creds.ID, creds.Email)
	if err != nil {
		return err
	}

	exchangeCode, err := h.storeOIDCExchange(c.Context(), tokenPair)
	if err != nil {
		return err
	}

	frontend := strings.TrimRight(h.FrontendURL, "/")
	if frontend == "" {
		frontend = "http://localhost:3000"
	}
	q := url.Values{}
	q.Set("code", exchangeCode)
	redirectURL := frontend + "/auth/callback?" + q.Encode()
	return c.Redirect(redirectURL, fiber.StatusFound)
}

// OIDCExchange swaps a one-time OIDC login code for access/refresh tokens.
func (h *Handler) OIDCExchange(c *fiber.Ctx) error {
	var body struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&body); err != nil || body.Code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "code is required")
	}
	key := oidcExchangePrefix + body.Code
	raw, err := h.Redis.GetDel(c.Context(), key).Bytes()
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired code")
	}
	var pair tokenPair
	if err := json.Unmarshal(raw, &pair); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired code")
	}
	return c.JSON(pair)
}

func (h *Handler) storeOIDCExchange(ctx context.Context, pair *tokenPair) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	code := base64.RawURLEncoding.EncodeToString(b)
	payload, err := json.Marshal(pair)
	if err != nil {
		return "", err
	}
	if err := h.Redis.Set(ctx, oidcExchangePrefix+code, payload, oidcExchangeTTL).Err(); err != nil {
		return "", err
	}
	return code, nil
}

func (h *Handler) provisionOIDCUser(c *fiber.Ctx, email string, claims map[string]any) (*repository.UserCredentials, error) {
	displayName := auth.DisplayNameFromClaims(claims, email)
	passwordHash, err := randomUnusablePasswordHash()
	if err != nil {
		return nil, err
	}

	userID := uuid.New()
	orgSlug, err := h.Repo.UniqueOrgSlug(c.Context(), slugify(strings.Split(email, "@")[0]))
	if err != nil {
		return nil, err
	}

	err = h.Repo.RegisterUser(c.Context(), repository.RegisterInput{
		UserID:       userID,
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
		OrgID:        uuid.New(),
		OrgSlug:      orgSlug,
		OrgName:      displayName + "'s Workspace",
		ProjectID:    uuid.New(),
	})
	if err != nil {
		return nil, fiber.NewError(fiber.StatusConflict, "could not provision user")
	}

	return &repository.UserCredentials{ID: userID, Email: email, PasswordHash: passwordHash}, nil
}

func randomUnusablePasswordHash() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "$oidc$" + base64.RawURLEncoding.EncodeToString(b), nil
}

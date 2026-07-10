package jobs

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/plexus/backend/internal/repository"
	"github.com/plexus/backend/internal/search"
)

const (
	TaskSendNotificationEmail = "notification:email"
	TaskIndexIssue            = "search:index_issue"
	TaskDeleteIssueIndex      = "search:delete_issue"
)

// EmailPayload is the task payload for email notifications.
type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// IndexIssuePayload is the task payload for search indexing.
type IndexIssuePayload struct {
	IssueID   string `json:"issue_id"`
	ProjectID string `json:"project_id"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}

type Server struct {
	server *asynq.Server
	mux    *asynq.ServeMux
	search *search.Client
	repo   *repository.Repo
}

func NewServer(redisURL string, searchClient *search.Client, repo *repository.Repo) *Server {
	opts, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		opts = asynq.RedisClientOpt{Addr: "localhost:6379"}
	}

	srv := asynq.NewServer(opts, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
	})

	s := &Server{
		server: srv,
		search: searchClient,
		repo:   repo,
	}

	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskSendNotificationEmail, handleEmailTask)
	mux.HandleFunc(TaskIndexIssue, s.handleIndexIssueTask)
	mux.HandleFunc(TaskDeleteIssueIndex, s.handleDeleteIssueTask)

	s.mux = mux
	return s
}

func (s *Server) Start() error {
	return s.server.Run(s.mux)
}

func (s *Server) Shutdown() {
	s.server.Shutdown()
}

func NewClient(redisURL string) *asynq.Client {
	opts, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		opts = asynq.RedisClientOpt{Addr: "localhost:6379"}
	}
	return asynq.NewClient(opts)
}

func handleEmailTask(ctx context.Context, t *asynq.Task) error {
	var p EmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	log.Printf("sending email to %s: %s", p.To, p.Subject)

	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return nil
	}

	body, err := json.Marshal(map[string]string{
		"from":    "Plexus <notifications@plexus.app>",
		"to":      p.To,
		"subject": p.Subject,
		"text":    p.Body,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend api error %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

func (s *Server) handleIndexIssueTask(ctx context.Context, t *asynq.Task) error {
	var p IndexIssuePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	issueID, err := uuid.Parse(p.IssueID)
	if err != nil {
		return err
	}

	issue, err := s.repo.GetIssueForSearch(ctx, issueID)
	if err != nil {
		return err
	}

	doc := search.IssueDocument{
		ID:          issue.ID.String(),
		ProjectID:   issue.ProjectID.String(),
		Number:      issue.Number,
		Title:       issue.Title,
		Description: nullStringVal(issue.Description),
		Priority:    issue.Priority,
		AssigneeName: nullStringVal(issue.AssigneeName),
		StatusName:  issue.StatusName,
		CreatedAt:   issue.CreatedAt.Unix(),
	}
	if err := s.search.IndexIssue(ctx, doc); err != nil {
		return err
	}
	log.Printf("indexed issue %s", p.IssueID)
	return nil
}

func (s *Server) handleDeleteIssueTask(ctx context.Context, t *asynq.Task) error {
	var p struct{ IssueID string `json:"issue_id"` }
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	if err := s.search.DeleteIssue(ctx, p.IssueID); err != nil {
		return err
	}
	log.Printf("deleted issue %s from search index", p.IssueID)
	return nil
}

func nullStringVal(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a minimal Meilisearch HTTP client.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

type IssueDocument struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	Number      int64  `json:"number"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Priority    string `json:"priority"`
	AssigneeName string `json:"assignee_name,omitempty"`
	StatusName  string `json:"status_name"`
	CreatedAt   int64  `json:"created_at"`
}

type SearchResult struct {
	Hits  []IssueDocument `json:"hits"`
	Total int             `json:"estimatedTotalHits"`
}

const issuesIndex = "issues"

func (c *Client) IndexIssue(ctx context.Context, doc IssueDocument) error {
	return c.addDocuments(ctx, issuesIndex, []IssueDocument{doc})
}

func (c *Client) DeleteIssue(ctx context.Context, issueID string) error {
	return c.deleteDocument(ctx, issuesIndex, issueID)
}

func (c *Client) Search(ctx context.Context, query string, projectID string, limit int) (*SearchResult, error) {
	body := map[string]interface{}{
		"q":     query,
		"limit": limit,
		"filter": fmt.Sprintf("project_id = %q", projectID),
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/indexes/%s/search", c.baseURL, issuesIndex),
		bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) EnsureIndexes(ctx context.Context) error {
	// Create issues index with filterable attributes
	_ = c.createIndex(ctx, issuesIndex, "id")
	_ = c.updateFilterableAttributes(ctx, issuesIndex, []string{"project_id", "priority", "status_name"})
	return nil
}

func (c *Client) addDocuments(ctx context.Context, index string, docs interface{}) error {
	data, err := json.Marshal(docs)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/indexes/%s/documents", c.baseURL, index),
		bytes.NewReader(data))
	if err != nil {
		return err
	}
	c.setHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("meilisearch error: %s", b)
	}
	return nil
}

func (c *Client) deleteDocument(ctx context.Context, index, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("%s/indexes/%s/documents/%s", c.baseURL, index, id), nil)
	if err != nil {
		return err
	}
	c.setHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *Client) createIndex(ctx context.Context, uid, primaryKey string) error {
	body := map[string]string{"uid": uid, "primaryKey": primaryKey}
	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/indexes", c.baseURL), bytes.NewReader(data))
	if err != nil {
		return err
	}
	c.setHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *Client) updateFilterableAttributes(ctx context.Context, index string, attrs []string) error {
	data, _ := json.Marshal(attrs)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		fmt.Sprintf("%s/indexes/%s/settings/filterable-attributes", c.baseURL, index),
		bytes.NewReader(data))
	if err != nil {
		return err
	}
	c.setHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
}

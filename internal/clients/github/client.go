package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"trpc-GitHub-agent/internal/domain"
)

const defaultBaseURL = "https://api.github.com"

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type Option func(*Client)

func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(baseURL, "/")
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func WithToken(token string) Option {
	return func(c *Client) {
		c.token = token
	}
}

func NewClient(opts ...Option) *Client {
	client := &Client{
		baseURL: defaultBaseURL,
		token:   os.Getenv("GITHUB_TOKEN"),
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(client)
	}
	return client
}

func (c *Client) SearchRepositories(ctx context.Context, query string, limit int) ([]domain.Repository, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 30 {
		limit = 30
	}
	endpoint := fmt.Sprintf("%s/search/repositories?q=%s&sort=stars&order=desc&per_page=%d", c.baseURL, url.QueryEscape(query), limit)
	var result searchResponse
	if err := c.getJSON(ctx, endpoint, &result); err != nil {
		return nil, err
	}
	repos := make([]domain.Repository, 0, len(result.Items))
	for _, item := range result.Items {
		raw, _ := json.Marshal(item)
		repos = append(repos, item.toDomain(string(raw)))
	}
	return repos, nil
}

func (c *Client) GetReadme(ctx context.Context, fullName string) (string, error) {
	endpoint := fmt.Sprintf("%s/repos/%s/readme", c.baseURL, fullName)
	var result readmeResponse
	if err := c.getJSON(ctx, endpoint, &result); err != nil {
		return "", err
	}
	if result.Encoding != "base64" {
		return result.Content, nil
	}
	cleaned := strings.ReplaceAll(result.Content, "\n", "")
	decoded, err := base64.StdEncoding.DecodeString(cleaned)
	if err != nil {
		return "", fmt.Errorf("decode README content: %w", err)
	}
	return string(decoded), nil
}

func (c *Client) GetTree(ctx context.Context, fullName, branch string) ([]TreeItem, error) {
	if branch == "" {
		branch = "main"
	}
	endpoint := fmt.Sprintf("%s/repos/%s/git/trees/%s?recursive=1", c.baseURL, fullName, url.PathEscape(branch))
	var result treeResponse
	if err := c.getJSON(ctx, endpoint, &result); err != nil {
		return nil, err
	}
	return result.Tree, nil
}

func (c *Client) CountIssuesByLabel(ctx context.Context, fullName, label string) (int, error) {
	endpoint := fmt.Sprintf("%s/search/issues?q=%s", c.baseURL, url.QueryEscape(fmt.Sprintf("repo:%s is:issue is:open label:%q", fullName, label)))
	var result issueSearchResponse
	if err := c.getJSON(ctx, endpoint, &result); err != nil {
		return 0, err
	}
	return result.TotalCount, nil
}

func (c *Client) SearchIssues(ctx context.Context, fullName, query string, limit int) ([]IssueItem, error) {
	if limit <= 0 {
		limit = 10
	}
	endpoint := fmt.Sprintf("%s/search/issues?q=%s&per_page=%d", c.baseURL, url.QueryEscape(fmt.Sprintf("repo:%s %s", fullName, query)), limit)
	var result issueSearchResponseWithItems
	if err := c.getJSON(ctx, endpoint, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (c *Client) SearchPullRequests(ctx context.Context, fullName, query string, limit int) ([]PullRequestItem, error) {
	if limit <= 0 {
		limit = 10
	}
	endpoint := fmt.Sprintf("%s/search/issues?q=%s&per_page=%d", c.baseURL, url.QueryEscape(fmt.Sprintf("repo:%s is:pr %s", fullName, query)), limit)
	var result issueSearchResponseWithPRs
	if err := c.getJSON(ctx, endpoint, &result); err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (c *Client) getJSON(ctx context.Context, endpoint string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "trpc-GitHub-agent")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return readErr
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return explainHTTPError(resp.StatusCode, body)
	}
	if len(body) == 0 {
		return errors.New("github returned an empty response")
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("decode github response: %w", err)
	}
	return nil
}

func explainHTTPError(status int, body []byte) error {
	var payload struct {
		Message string `json:"message"`
	}
	_ = json.Unmarshal(body, &payload)
	message := strings.TrimSpace(payload.Message)
	if message == "" {
		message = http.StatusText(status)
	}
	switch status {
	case http.StatusForbidden:
		return fmt.Errorf("github request forbidden or rate limited: %s", message)
	case http.StatusNotFound:
		return fmt.Errorf("github repository or resource not found: %s", message)
	default:
		return fmt.Errorf("github request failed with status %d: %s", status, message)
	}
}

type searchResponse struct {
	Items []repositoryItem `json:"items"`
}

type repositoryItem struct {
	ID              int64     `json:"id"`
	FullName        string    `json:"full_name"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	HTMLURL         string    `json:"html_url"`
	CloneURL        string    `json:"clone_url"`
	Language        string    `json:"language"`
	Topics          []string  `json:"topics"`
	Stars           int       `json:"stargazers_count"`
	Forks           int       `json:"forks_count"`
	Watchers        int       `json:"watchers_count"`
	OpenIssuesCount int       `json:"open_issues_count"`
	DefaultBranch   string    `json:"default_branch"`
	Archived        bool      `json:"archived"`
	Disabled        bool      `json:"disabled"`
	PushedAt        time.Time `json:"pushed_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Owner           struct {
		Login string `json:"login"`
	} `json:"owner"`
}

func (r repositoryItem) toDomain(raw string) domain.Repository {
	return domain.Repository{
		GitHubID:        r.ID,
		FullName:        r.FullName,
		Owner:           r.Owner.Login,
		Name:            r.Name,
		Description:     r.Description,
		HTMLURL:         r.HTMLURL,
		CloneURL:        r.CloneURL,
		Language:        r.Language,
		Topics:          r.Topics,
		Stars:           r.Stars,
		Forks:           r.Forks,
		Watchers:        r.Watchers,
		OpenIssuesCount: r.OpenIssuesCount,
		DefaultBranch:   r.DefaultBranch,
		Archived:        r.Archived,
		Disabled:        r.Disabled,
		PushedAt:        r.PushedAt,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
		FetchedAt:       time.Now(),
		RawJSON:         raw,
	}
}

type readmeResponse struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

type TreeItem struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type treeResponse struct {
	Tree []TreeItem `json:"tree"`
}

type issueSearchResponse struct {
	TotalCount int `json:"total_count"`
}

type issueSearchResponseWithItems struct {
	TotalCount int         `json:"total_count"`
	Items      []IssueItem `json:"items"`
}

type issueSearchResponseWithPRs struct {
	TotalCount int              `json:"total_count"`
	Items      []PullRequestItem `json:"items"`
}

type IssueItem struct {
	Title     string   `json:"title"`
	HTMLURL   string   `json:"html_url"`
	Labels    []Label  `json:"labels"`
	State     string   `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PullRequestItem struct {
	Title     string   `json:"title"`
	HTMLURL   string   `json:"html_url"`
	Labels    []Label  `json:"labels"`
	State     string   `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Label struct {
	Name string `json:"name"`
}

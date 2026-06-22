package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchRepositoriesAddsTokenHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("expected authorization header, got %q", got)
		}
		if got := r.Header.Get("X-GitHub-Api-Version"); got != "2022-11-28" {
			t.Fatalf("expected API version header, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"id":1,"full_name":"owner/repo","name":"repo","description":"Agent repo","html_url":"https://github.com/owner/repo","clone_url":"https://github.com/owner/repo.git","language":"Go","topics":["agent"],"stargazers_count":100,"forks_count":10,"watchers_count":100,"open_issues_count":3,"default_branch":"main","archived":false,"disabled":false,"pushed_at":"2026-06-01T00:00:00Z","created_at":"2025-01-01T00:00:00Z","updated_at":"2026-06-01T00:00:00Z","owner":{"login":"owner"}}]}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL), WithToken("test-token"))
	repos, err := client.SearchRepositories(context.Background(), "agent language:Go", 1)
	if err != nil {
		t.Fatalf("search repositories: %v", err)
	}
	if len(repos) != 1 || repos[0].FullName != "owner/repo" {
		t.Fatalf("unexpected repos: %#v", repos)
	}
}

func TestGitHubClientExplainsRateLimitError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"API rate limit exceeded"}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	_, err := client.SearchRepositories(context.Background(), "agent", 1)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "rate limited") {
		t.Fatalf("expected rate limit explanation, got %v", err)
	}
}

func TestGitHubClientExplainsNotFoundError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	_, err := client.GetReadme(context.Background(), "owner/missing")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found explanation, got %v", err)
	}
}

func TestGitHubClientSearchIssuesAndPRs(t *testing.T) {
	issueServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "repo%3Aowner%2Frepo") {
			t.Fatalf("expected repo-qualified query, got %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total_count":1,"items":[{"title":"Add docs","html_url":"https://github.com/owner/repo/issues/1","labels":[{"name":"docs"}],"state":"open","created_at":"2026-06-01T00:00:00Z","updated_at":"2026-06-02T00:00:00Z"}]}`))
	}))
	defer issueServer.Close()

	client := NewClient(WithBaseURL(issueServer.URL))
	issues, err := client.SearchIssues(context.Background(), "owner/repo", "label:docs", 1)
	if err != nil {
		t.Fatalf("search issues: %v", err)
	}
	if len(issues) != 1 || issues[0].Title != "Add docs" {
		t.Fatalf("unexpected issues: %#v", issues)
	}

	prServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "is%3Apr") {
			t.Fatalf("expected PR query, got %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total_count":1,"items":[{"title":"Refactor parser","html_url":"https://github.com/owner/repo/pull/2","labels":[{"name":"refactor"}],"state":"open","created_at":"2026-06-01T00:00:00Z","updated_at":"2026-06-02T00:00:00Z"}]}`))
	}))
	defer prServer.Close()

	prClient := NewClient(WithBaseURL(prServer.URL))
	prs, err := prClient.SearchPullRequests(context.Background(), "owner/repo", "is:open", 1)
	if err != nil {
		t.Fatalf("search pull requests: %v", err)
	}
	if len(prs) != 1 || prs[0].Title != "Refactor parser" {
		t.Fatalf("unexpected prs: %#v", prs)
	}
}

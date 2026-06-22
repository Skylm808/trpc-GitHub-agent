package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"trpc-GitHub-agent/internal/domain"
)

func TestSQLiteStoreWritesAndReadsRepository(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "research.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	repo := domain.Repository{
		GitHubID:        42,
		FullName:        "owner/repo",
		Owner:           "owner",
		Name:            "repo",
		Description:     "Go agent repo",
		HTMLURL:         "https://github.com/owner/repo",
		CloneURL:        "https://github.com/owner/repo.git",
		Language:        "Go",
		Topics:          []string{"agent", "rag"},
		Stars:           120,
		Forks:           12,
		Watchers:        120,
		OpenIssuesCount: 4,
		DefaultBranch:   "main",
		PushedAt:        time.Now().UTC(),
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	id, err := store.UpsertRepository(context.Background(), repo)
	if err != nil {
		t.Fatalf("upsert repo: %v", err)
	}
	if id == 0 {
		t.Fatal("expected repository id")
	}

	got, ok, err := store.GetRepositoryByFullName(context.Background(), "owner/repo")
	if err != nil {
		t.Fatalf("get repo: %v", err)
	}
	if !ok {
		t.Fatal("expected cached repository")
	}
	if got.FullName != repo.FullName || got.Topics[1] != "rag" {
		t.Fatalf("unexpected repository: %#v", got)
	}
}

func TestSQLiteStoreSavesScoreReportAndQueryHistory(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "research.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	repoID, err := store.UpsertRepository(context.Background(), domain.Repository{
		GitHubID:  7,
		FullName:  "owner/repo",
		Owner:     "owner",
		Name:      "repo",
		PushedAt:  time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		FetchedAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("upsert repo: %v", err)
	}

	score := domain.Score{
		RepositoryID: repoID,
		TotalScore:   80,
		Explanation:  map[string]string{"activity": "ok"},
		ScoredAt:     time.Now().UTC(),
	}
	if err := store.SaveScore(context.Background(), score); err != nil {
		t.Fatalf("save score: %v", err)
	}
	if err := store.SaveReport(context.Background(), repoID, "recommendation", "Report", "# Report"); err != nil {
		t.Fatalf("save report: %v", err)
	}
	if err := store.SaveQueryHistory(context.Background(), domain.SearchIntent{UserInput: "go agent"}, nil, []int64{repoID}); err != nil {
		t.Fatalf("save query history: %v", err)
	}
}

func TestSQLiteStoreSavesAndReadsResearchSession(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "research.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	analysis := domain.RepositoryAnalysis{
		Repository:   domain.Repository{FullName: "owner/repo", Owner: "owner", Name: "repo"},
		IssueSummary: "docs: 1, bug: 1",
		LLMInsight: domain.LLMInsight{
			Provider:    "custom",
			Model:       "gpt-compatible",
			AIGenerated: true,
		},
		AgentTrace: []domain.AgentTraceStep{{Phase: "Plan", Tool: "analyze_repository", Summary: "ok"}},
	}
	id, err := store.SaveResearchSession(context.Background(), analysis)
	if err != nil {
		t.Fatalf("save research session: %v", err)
	}
	if id == 0 {
		t.Fatal("expected session id")
	}

	sessions, err := store.ListResearchSessions(context.Background(), 10)
	if err != nil {
		t.Fatalf("list research sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("expected one session, got %d", len(sessions))
	}
	if sessions[0].Repository != "owner/repo" || !sessions[0].AIGenerated || sessions[0].TraceStepCount != 1 {
		t.Fatalf("unexpected session summary: %#v", sessions[0])
	}

	got, ok, err := store.GetResearchSession(context.Background(), id)
	if err != nil {
		t.Fatalf("get research session: %v", err)
	}
	if !ok {
		t.Fatal("expected saved session")
	}
	if got.Analysis.Repository.FullName != "owner/repo" || got.Analysis.IssueSummary == "" {
		t.Fatalf("unexpected session analysis: %#v", got.Analysis)
	}
}

func TestAnalyzeTreeDetectsProjectSignals(t *testing.T) {
	hasDocs, hasExamples, hasTests, hasContributing, dependencyFiles, dependencySummary, structureSummary := AnalyzeTree([]string{
		"README.md", "go.mod", "docs/index.md", "examples/basic/main.go", "internal/app/app_test.go", "CONTRIBUTING.md",
	})

	if !hasDocs || !hasExamples || !hasTests || !hasContributing {
		t.Fatalf("expected project signals, got docs=%v examples=%v tests=%v contributing=%v", hasDocs, hasExamples, hasTests, hasContributing)
	}
	if dependencySummary == "" || structureSummary == "" {
		t.Fatalf("expected summaries, got %q / %q", dependencySummary, structureSummary)
	}
	if len(dependencyFiles) != 1 || dependencyFiles[0] != "go.mod" {
		t.Fatalf("expected dependency files, got %#v", dependencyFiles)
	}
}

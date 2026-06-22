package queryplanner

import (
	"strings"
	"testing"

	"trpc-GitHub-agent/internal/domain"
)

func TestPlannerGeneratesGoAgentContributionQueries(t *testing.T) {
	planner := NewPlanner()
	intent, queries := planner.Plan("我是 Go 后端，帮我找 Go Agent 项目，适合秋招和开源贡献", 5, domain.SearchIntent{
		InputLanguage: "zh",
		Direction:     "mcp",
		PushedAfter:   "2025-05-01",
		MinStars:      500,
		MaxStars:      20000,
		Difficulty:    "intermediate",
		TargetRole:    "backend",
		Languages:     []string{"Go"},
		Topics:        []string{"agent"},
	})

	if len(intent.Languages) != 1 || intent.Languages[0] != "Go" {
		t.Fatalf("expected Go language, got %#v", intent.Languages)
	}
	if intent.TargetRole != "backend" {
		t.Fatalf("expected backend role, got %q", intent.TargetRole)
	}
	if len(queries) == 0 {
		t.Fatal("expected generated queries")
	}
	joined := strings.Join([]string{queries[0].Query, queries[len(queries)-1].Query}, " ")
	for _, want := range []string{"language:Go", "archived:false", "pushed:>2025-05-01"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected query to contain %q, got %q", want, joined)
		}
	}
	if !strings.Contains(joined, "stars:500..20000") {
		t.Fatalf("expected star range in query, got %q", joined)
	}
	if !strings.Contains(joined, "topic:mcp") {
		t.Fatalf("expected direction filter in query, got %#v", queries)
	}
	if !strings.Contains(joined, "good-first-issues:>0") {
		t.Fatalf("expected contribution query, got %#v", queries)
	}
}

func TestPlannerGeneratesPythonRAGQuery(t *testing.T) {
	planner := NewPlanner()
	intent, queries := planner.Plan("I use Python and want Python RAG projects for resume building", 10, domain.SearchIntent{})

	if intent.Languages[0] != "Python" {
		t.Fatalf("expected Python language, got %#v", intent.Languages)
	}
	foundRAG := false
	for _, query := range queries {
		if strings.Contains(query.Query, "rag OR retrieval") && strings.Contains(query.Query, "language:Python") {
			foundRAG = true
		}
	}
	if !foundRAG {
		t.Fatalf("expected Python RAG query, got %#v", queries)
	}
}

package queryplanner

import (
	"strings"
	"testing"
)

func TestPlannerGeneratesGoAgentContributionQueries(t *testing.T) {
	planner := NewPlanner()
	intent, queries := planner.Plan("我是 Go 后端，帮我找 Go Agent 项目，适合秋招和开源贡献", 5)

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
	for _, want := range []string{"language:Go", "archived:false", "pushed:>2025-01-01"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected query to contain %q, got %q", want, joined)
		}
	}
	if !strings.Contains(joined, "good-first-issues:>0") {
		t.Fatalf("expected contribution query, got %#v", queries)
	}
}

func TestPlannerGeneratesPythonRAGQuery(t *testing.T) {
	planner := NewPlanner()
	intent, queries := planner.Plan("I use Python and want Python RAG projects for resume building", 10)

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

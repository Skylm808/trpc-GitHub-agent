package main

import (
	"testing"

	researchagent "trpc-GitHub-agent/internal/agent"
	appsvc "trpc-GitHub-agent/internal/app"
	"trpc-GitHub-agent/internal/domain"
)

func TestDiscoverProjectsFallbackCoreFlow(t *testing.T) {
	app := NewApp()
	app.store = nil
	app.discovery = appsvc.NewDiscoveryService(nil, nil)
	app.agent = researchagent.NewRunner(app.discovery, nil)

	result, err := app.DiscoverProjects(domain.SearchRequest{
		UserInput:  "我是 Go 后端，帮我找 Go Agent 项目，适合秋招和开源贡献。",
		Limit:      5,
		MinStars:   100,
		Direction:  "agent",
		Difficulty: "beginner",
	})
	if err != nil {
		t.Fatalf("discover projects: %v", err)
	}
	if len(result.Queries) == 0 {
		t.Fatal("expected generated queries")
	}
	if len(result.Repositories) == 0 {
		t.Fatal("expected scored repositories")
	}
	if result.Repositories[0].Score.TotalScore == 0 {
		t.Fatalf("expected scored first repository: %#v", result.Repositories[0])
	}
	if result.MarkdownReport == "" {
		t.Fatal("expected markdown report")
	}
	if result.UsedLiveGitHub {
		t.Fatal("expected fallback mode")
	}
}

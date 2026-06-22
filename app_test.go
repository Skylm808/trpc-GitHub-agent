package main

import (
	"strconv"
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

func TestRepositoryHistoryMessagesKeepsRecentTurns(t *testing.T) {
	var history []domain.RepositoryQuestionTurn
	for i := 0; i < 8; i++ {
		history = append(history, domain.RepositoryQuestionTurn{
			Question: "q" + strconv.Itoa(i),
			Answer:   "a" + strconv.Itoa(i),
		})
	}
	messages := repositoryHistoryMessages(history)
	if len(messages) != 12 {
		t.Fatalf("expected six turns as twelve messages, got %d", len(messages))
	}
	if messages[0].Content != "历史问题：q2" || messages[1].Content != "a2" {
		t.Fatalf("expected recent history to start at q2/a2, got %#v %#v", messages[0], messages[1])
	}
	if messages[len(messages)-2].Content != "历史问题：q7" || messages[len(messages)-1].Content != "a7" {
		t.Fatalf("expected history to end at q7/a7, got %#v %#v", messages[len(messages)-2], messages[len(messages)-1])
	}
}

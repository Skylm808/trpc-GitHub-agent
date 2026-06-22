package agent

import (
	"context"
	"testing"

	appsvc "trpc-GitHub-agent/internal/app"
)

func TestToolsetRegistersCoreTools(t *testing.T) {
	toolset := NewToolset(appsvc.NewDiscoveryService(nil, nil), nil)
	names := map[string]bool{}
	for _, tl := range toolset.Tools() {
		declaration := tl.Declaration()
		if declaration != nil {
			names[declaration.Name] = true
		}
	}
	for _, want := range []string{
		"search_repositories",
		"score_repository",
		"generate_project_report",
		"remember_user_preference",
	} {
		if !names[want] {
			t.Fatalf("expected tool %s to be registered, got %#v", want, names)
		}
	}
}

func TestRunnerDiscoverProjectsUsesFrameworkPath(t *testing.T) {
	discovery := appsvc.NewDiscoveryService(nil, nil)
	runner := NewRunner(discovery, nil)

	result, err := runner.DiscoverProjects(context.Background(), "我是 Go 后端，想找 Agent 开源项目", 3)
	if err != nil {
		t.Fatalf("discover projects through agent runner: %v", err)
	}
	if len(result.Repositories) == 0 {
		t.Fatal("expected repositories from deterministic fallback")
	}
	if len(result.Queries) == 0 {
		t.Fatal("expected generated queries")
	}
}

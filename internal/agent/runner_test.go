package agent

import (
	"context"
	"testing"

	appsvc "trpc-GitHub-agent/internal/app"
	"trpc-GitHub-agent/internal/domain"
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
		"get_repository_metadata",
		"get_readme",
		"get_tree",
		"get_dependency_files",
		"classify_issues",
		"summarize_prs",
		"score_repository",
		"generate_contribution_plan",
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
	if len(result.AgentTrace) == 0 {
		t.Fatal("expected discovery agent trace")
	}
}

func TestRunnerDiscoverProjectsWithRequestUsesFrameworkPath(t *testing.T) {
	discovery := appsvc.NewDiscoveryService(nil, nil)
	runner := NewRunner(discovery, nil)

	result, err := runner.DiscoverProjectsWithRequest(context.Background(), domain.SearchRequest{
		UserInput: "我是 Go 后端，想找 Agent 开源项目",
		Limit:     3,
	})
	if err != nil {
		t.Fatalf("discover projects with request: %v", err)
	}
	if len(result.Repositories) == 0 {
		t.Fatal("expected repositories from deterministic fallback")
	}
	seen := map[string]bool{}
	for _, step := range result.AgentTrace {
		seen[step.Tool] = true
	}
	for _, want := range []string{"plan_search_queries", "search_repositories", "score_repository", "generate_project_report"} {
		if !seen[want] {
			t.Fatalf("expected discovery trace tool %s, got %#v", want, result.AgentTrace)
		}
	}
}

func TestRunnerAnalyzeRepositoryUsesToolChain(t *testing.T) {
	discovery := appsvc.NewDiscoveryService(nil, nil)
	runner := NewRunner(discovery, nil)

	analysis, err := runner.AnalyzeRepository(context.Background(), "trpc-group/trpc-agent-go")
	if err != nil {
		t.Fatalf("analyze repository through agent runner: %v", err)
	}
	if analysis.Repository.FullName != "trpc-group/trpc-agent-go" {
		t.Fatalf("unexpected repository: %#v", analysis.Repository)
	}
	if analysis.ContributionPlan == "" || analysis.ResumeValue == "" {
		t.Fatalf("expected contribution plan and resume value: %#v", analysis)
	}
	if analysis.Positioning == "" || analysis.Architecture == "" {
		t.Fatalf("expected positioning and architecture: %#v", analysis)
	}
	if len(analysis.LearningModules) == 0 || len(analysis.ContributionTypes) == 0 {
		t.Fatalf("expected structured modules and contribution types: %#v", analysis)
	}
	seen := map[string]bool{}
	for _, step := range analysis.AgentTrace {
		seen[step.Tool] = true
	}
	for _, want := range []string{
		"get_repository_metadata",
		"get_readme",
		"get_tree",
		"get_dependency_files",
		"classify_issues",
		"summarize_prs",
		"score_repository",
		"generate_contribution_plan",
	} {
		if !seen[want] {
			t.Fatalf("expected trace tool %s, got %#v", want, analysis.AgentTrace)
		}
	}
}

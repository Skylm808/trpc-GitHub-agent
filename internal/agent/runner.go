package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	appsvc "trpc-GitHub-agent/internal/app"
	"trpc-GitHub-agent/internal/domain"
	"trpc-GitHub-agent/internal/store/sqlite"

	coreagent "trpc.group/trpc-go/trpc-agent-go/agent"
	"trpc.group/trpc-go/trpc-agent-go/event"
	"trpc.group/trpc-go/trpc-agent-go/model"
	trpcrunner "trpc.group/trpc-go/trpc-agent-go/runner"
	"trpc.group/trpc-go/trpc-agent-go/tool"
)

const (
	researchAgentName = "open_source_project_researcher"
	appName           = "trpc-github-agent"
)

// Runner 通过 tRPC-Agent-Go runner 执行项目研究工作流。
type Runner struct {
	framework trpcrunner.Runner
	toolset   *Toolset
}

// NewRunner 创建一个确定性 Agent runner；后续可在这里切换到 LLM Agent。
func NewRunner(discovery *appsvc.DiscoveryService, store *sqlite.SQLiteStore) *Runner {
	toolset := NewToolset(discovery, store)
	agent := &researchAgent{
		name:  researchAgentName,
		tools: toolset.Tools(),
	}
	return &Runner{
		framework: trpcrunner.NewRunner(appName, agent),
		toolset:   toolset,
	}
}

// DiscoverProjects 通过框架 runner 执行一次项目发现，并反序列化 Agent 返回的结构化结果。
func (r *Runner) DiscoverProjects(ctx context.Context, userInput string, limit int) (domain.DiscoveryResult, error) {
	return r.DiscoverProjectsWithRequest(ctx, domain.SearchRequest{UserInput: userInput, Limit: limit})
}

// DiscoverProjectsWithRequest 通过框架 runner 执行一次结构化项目发现。
func (r *Runner) DiscoverProjectsWithRequest(ctx context.Context, request domain.SearchRequest) (domain.DiscoveryResult, error) {
	if r == nil || r.framework == nil {
		return domain.DiscoveryResult{}, fmt.Errorf("agent runner is not configured")
	}
	payload, err := json.Marshal(request)
	if err != nil {
		return domain.DiscoveryResult{}, err
	}
	events, err := r.framework.Run(ctx, "local-user", "default-session", model.NewUserMessage(string(payload)))
	if err != nil {
		return domain.DiscoveryResult{}, err
	}

	var content string
	for evt := range events {
		if evt == nil || evt.Response == nil {
			continue
		}
		if evt.Error != nil {
			return domain.DiscoveryResult{}, fmt.Errorf("agent run failed: %s", evt.Error.Message)
		}
		if len(evt.Choices) > 0 {
			content = evt.Choices[0].Message.Content
		}
	}
	if strings.TrimSpace(content) == "" {
		return domain.DiscoveryResult{}, fmt.Errorf("agent returned an empty result")
	}
	var result domain.DiscoveryResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return domain.DiscoveryResult{}, fmt.Errorf("decode agent discovery result: %w", err)
	}
	return result, nil
}

// AnalyzeRepository 使用显式工具链执行单仓库研究流程。
func (r *Runner) AnalyzeRepository(ctx context.Context, fullName string) (domain.RepositoryAnalysis, error) {
	if r == nil || r.toolset == nil {
		return domain.RepositoryAnalysis{}, fmt.Errorf("agent runner is not configured")
	}
	if strings.TrimSpace(fullName) == "" {
		return domain.RepositoryAnalysis{}, fmt.Errorf("repository full_name is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	trace := []domain.AgentTraceStep{
		{Phase: "Plan", Tool: "AnalyzeRepository", Summary: "制定只读研究流程：元数据 -> README -> 目录树 -> 依赖 -> Issue/PR -> 评分 -> 贡献计划。"},
	}

	repo, err := r.toolset.getRepositoryMetadata(ctx, RepositoryInput{FullName: fullName})
	if err != nil {
		return domain.RepositoryAnalysis{}, err
	}
	trace = append(trace, domain.AgentTraceStep{Phase: "Tool Calls", Tool: "get_repository_metadata", Summary: summarizeRepositoryMetadata(repo)})

	readmeSummary, err := r.toolset.getReadme(ctx, ReadmeInput{Repository: repo})
	if err != nil {
		return domain.RepositoryAnalysis{}, err
	}
	trace = append(trace, domain.AgentTraceStep{Phase: "Tool Calls", Tool: "get_readme", Summary: readmeSummary})

	tree, err := r.toolset.getTree(ctx, TreeInput{Repository: repo})
	if err != nil {
		return domain.RepositoryAnalysis{}, err
	}
	trace = append(trace, domain.AgentTraceStep{Phase: "Tool Calls", Tool: "get_tree", Summary: tree.DirectorySummary})
	trace = append(trace, domain.AgentTraceStep{Phase: "Tool Calls", Tool: "get_dependency_files", Summary: summarizeDependencies(tree.DependencyFiles, tree.DependencySummary)})

	issueSummary, err := r.toolset.classifyIssues(ctx, RepositoryInput{FullName: fullName})
	if err != nil {
		return domain.RepositoryAnalysis{}, err
	}
	trace = append(trace, domain.AgentTraceStep{Phase: "Tool Calls", Tool: "classify_issues", Summary: issueSummary})

	prSummary, err := r.toolset.summarizePRs(ctx, RepositoryInput{FullName: fullName})
	if err != nil {
		return domain.RepositoryAnalysis{}, err
	}
	trace = append(trace, domain.AgentTraceStep{Phase: "Tool Calls", Tool: "summarize_prs", Summary: prSummary})

	profile := domain.RepositoryProfile{
		Repository:        repo,
		ReadmeSummary:     readmeSummary,
		StructureSummary:  tree.DirectorySummary,
		DependencySummary: tree.DependencySummary,
		HasReadme:         strings.TrimSpace(readmeSummary) != "",
		HasDocs:           strings.Contains(strings.ToLower(tree.DirectorySummary), "docs"),
		HasExamples:       strings.Contains(strings.ToLower(tree.DirectorySummary), "examples"),
		HasTests:          strings.Contains(strings.ToLower(tree.DirectorySummary), "test"),
	}
	score, err := r.toolset.scoreRepository(ctx, ScoreRepositoryInput{Profile: profile, Intent: domain.SearchIntent{UserInput: fullName}})
	if err != nil {
		return domain.RepositoryAnalysis{}, err
	}
	trace = append(trace, domain.AgentTraceStep{Phase: "Tool Calls", Tool: "score_repository", Summary: fmt.Sprintf("确定性评分完成：%d/100，影响力 %s，难度 %s。", score.TotalScore, score.InfluenceLevel, score.Difficulty)})

	analysis := domain.RepositoryAnalysis{
		Repository:        repo,
		Profile:           profile,
		Positioning:       summarizeRunnerPositioning(repo),
		Architecture:      tree.DirectorySummary,
		LearningModules:   runnerLearningModules(profile, tree.DependencyFiles),
		ContributionTypes: runnerContributionTypes(issueSummary, profile),
		IssueSummary:      issueSummary,
		PRSummary:         prSummary,
		DocsSummary:       signalFromText(tree.DirectorySummary, "docs", "检测到 docs 目录或文档文件"),
		ExamplesSummary:   signalFromText(tree.DirectorySummary, "examples", "检测到 examples 示例目录"),
		TestsSummary:      signalFromText(tree.DirectorySummary, "test", "检测到测试目录或测试文件"),
		DependencyFiles:   tree.DependencyFiles,
		DirectorySummary:  tree.DirectorySummary,
	}
	plan, err := r.toolset.generateContributionPlan(ctx, ContributionPlanInput{Analysis: analysis})
	if err != nil {
		return domain.RepositoryAnalysis{}, err
	}
	analysis.ContributionPlan = plan.ContributionPlan
	analysis.ResumeValue = plan.ResumeValue
	analysis.AgentTrace = append(trace, domain.AgentTraceStep{Phase: "Findings", Tool: "generate_contribution_plan", Summary: "生成 7 天贡献路线和简历/面试价值总结。"})
	return analysis, nil
}

type researchAgent struct {
	name  string
	tools []tool.Tool
}

func (a *researchAgent) Info() coreagent.Info {
	return coreagent.Info{
		Name:        a.name,
		Description: "GitHub open-source project researcher for recruiting and learning workflows.",
	}
}

func (a *researchAgent) Tools() []tool.Tool {
	return a.tools
}

func (a *researchAgent) SubAgents() []coreagent.Agent {
	return nil
}

func (a *researchAgent) FindSubAgent(name string) coreagent.Agent {
	return nil
}

func (a *researchAgent) Run(ctx context.Context, invocation *coreagent.Invocation) (<-chan *event.Event, error) {
	ch := make(chan *event.Event, 1)
	go func() {
		defer close(ch)
		if invocation == nil {
			ch <- event.NewErrorEvent("", a.name, model.ErrorTypeFlowError, "nil invocation")
			return
		}
		result, err := a.callSearchTool(ctx, []byte(invocation.Message.Content))
		if err != nil {
			ch <- event.NewErrorEvent(invocation.InvocationID, a.name, model.ErrorTypeFlowError, err.Error())
			return
		}
		body, err := json.Marshal(result)
		if err != nil {
			ch <- event.NewErrorEvent(invocation.InvocationID, a.name, model.ErrorTypeFlowError, err.Error())
			return
		}
		ch <- event.NewResponseEvent(invocation.InvocationID, a.name, &model.Response{
			Object: model.ObjectTypeChatCompletion,
			Done:   true,
			Choices: []model.Choice{
				{Index: 0, Message: model.NewAssistantMessage(string(body))},
			},
		})
	}()
	return ch, nil
}

func (a *researchAgent) callSearchTool(ctx context.Context, args []byte) (domain.DiscoveryResult, error) {
	for _, candidate := range a.tools {
		declaration := candidate.Declaration()
		if declaration == nil || declaration.Name != "search_repositories" {
			continue
		}
		callable, ok := candidate.(tool.CallableTool)
		if !ok {
			return domain.DiscoveryResult{}, fmt.Errorf("search_repositories is not callable")
		}
		result, err := callable.Call(ctx, args)
		if err != nil {
			return domain.DiscoveryResult{}, err
		}
		discoveryResult, ok := result.(domain.DiscoveryResult)
		if !ok {
			return domain.DiscoveryResult{}, fmt.Errorf("unexpected search_repositories result type %T", result)
		}
		return discoveryResult, nil
	}
	return domain.DiscoveryResult{}, fmt.Errorf("search_repositories tool is not registered")
}

func summarizeRepositoryMetadata(repo domain.Repository) string {
	return fmt.Sprintf("%s：%s，语言 %s，stars %d，forks %d，open issues %d。",
		repo.FullName, repo.Description, repo.Language, repo.Stars, repo.Forks, repo.OpenIssuesCount)
}

func summarizeDependencies(files []string, summary string) string {
	if summary != "" {
		return summary
	}
	if len(files) == 0 {
		return "未识别到依赖文件。"
	}
	return "Dependency files: " + strings.Join(files, ", ")
}

func signalFromText(text, token, positive string) string {
	if strings.Contains(strings.ToLower(text), token) {
		return positive
	}
	return "未检测到该信号"
}

func summarizeRunnerPositioning(repo domain.Repository) string {
	var parts []string
	if strings.TrimSpace(repo.Language) != "" {
		parts = append(parts, repo.Language+" 项目")
	}
	for _, topic := range repo.Topics {
		if strings.TrimSpace(topic) != "" {
			parts = append(parts, topic)
		}
		if len(parts) >= 4 {
			break
		}
	}
	if len(parts) == 0 {
		parts = append(parts, "开源项目")
	}
	description := strings.TrimSpace(repo.Description)
	if description == "" {
		description = "适合从 README、目录树、Issue 和 PR 信号继续判断贡献价值。"
	}
	return repo.FullName + " 定位为 " + strings.Join(parts, " / ") + "；" + description
}

func runnerLearningModules(profile domain.RepositoryProfile, dependencyFiles []string) []string {
	modules := []string{"README / Quickstart"}
	if profile.HasDocs {
		modules = append(modules, "docs")
	}
	if profile.HasExamples {
		modules = append(modules, "examples")
	}
	if profile.HasTests {
		modules = append(modules, "tests")
	}
	if len(dependencyFiles) > 0 {
		modules = append(modules, "dependency files: "+strings.Join(dependencyFiles, ", "))
	}
	if len(modules) == 1 {
		modules = append(modules, "core source directories")
	}
	return modules
}

func runnerContributionTypes(issueSummary string, profile domain.RepositoryProfile) []string {
	lower := strings.ToLower(issueSummary)
	seen := map[string]bool{}
	var result []string
	for _, item := range []string{"docs", "bug", "feature", "test", "good-first", "infra"} {
		if strings.Contains(lower, item) {
			seen[item] = true
			result = append(result, item)
		}
	}
	if strings.Contains(issueSummary, "文档") && !seen["docs"] {
		result = append(result, "docs")
	}
	if strings.Contains(issueSummary, "缺陷") && !seen["bug"] {
		result = append(result, "bug")
	}
	if strings.Contains(issueSummary, "功能") && !seen["feature"] {
		result = append(result, "feature")
	}
	if strings.Contains(issueSummary, "测试") && !seen["test"] {
		result = append(result, "test")
	}
	if profile.GoodFirstIssueCount > 0 && !seen["good-first"] {
		result = append(result, "good-first")
	}
	if len(result) == 0 {
		result = append(result, "docs", "test", "good-first")
	}
	return result
}

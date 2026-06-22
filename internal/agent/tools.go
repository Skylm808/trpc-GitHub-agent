package agent

import (
	"context"
	"errors"

	appsvc "trpc-GitHub-agent/internal/app"
	"trpc-GitHub-agent/internal/domain"
	"trpc-GitHub-agent/internal/services/report"
	"trpc-GitHub-agent/internal/services/scoring"
	"trpc-GitHub-agent/internal/store/sqlite"

	"trpc.group/trpc-go/trpc-agent-go/tool"
	"trpc.group/trpc-go/trpc-agent-go/tool/function"
)

// SearchRepositoriesInput 是 search_repositories tool 的输入。
type SearchRepositoriesInput = domain.SearchRequest

// ScoreRepositoryInput 是 score_repository tool 的输入。
type ScoreRepositoryInput struct {
	Profile domain.RepositoryProfile `json:"profile" jsonschema:"description=仓库画像"`
	Intent  domain.SearchIntent      `json:"intent" jsonschema:"description=用户搜索意图"`
}

// RepositoryInput 是单仓库只读分析工具的通用输入。
type RepositoryInput struct {
	FullName string `json:"full_name" jsonschema:"description=GitHub repository full_name, e.g. trpc-group/trpc-agent-go"`
}

// ReadmeInput 是 README 摘要工具输入。
type ReadmeInput struct {
	Repository domain.Repository `json:"repository" jsonschema:"description=仓库基础元数据"`
}

// TreeInput 是目录树和依赖文件工具输入。
type TreeInput struct {
	Repository domain.Repository `json:"repository" jsonschema:"description=仓库基础元数据"`
}

// TreeResult 是 get_tree 和 get_dependency_files 的共享结果。
type TreeResult struct {
	DirectorySummary  string   `json:"directory_summary"`
	DependencyFiles   []string `json:"dependency_files"`
	DependencySummary string   `json:"dependency_summary"`
}

// ContributionPlanInput 是贡献路线生成工具输入。
type ContributionPlanInput struct {
	Analysis domain.RepositoryAnalysis `json:"analysis" jsonschema:"description=确定性单仓库分析结果"`
}

// ContributionPlanResult 是贡献计划和简历价值结果。
type ContributionPlanResult struct {
	ContributionPlan string `json:"contribution_plan"`
	ResumeValue      string `json:"resume_value"`
}

// GenerateProjectReportInput 是 generate_project_report tool 的输入。
type GenerateProjectReportInput struct {
	Result domain.DiscoveryResult `json:"result" jsonschema:"description=项目发现结果"`
}

// RememberUserPreferenceInput 是 remember_user_preference tool 的输入。
type RememberUserPreferenceInput struct {
	Key   string `json:"key" jsonschema:"description=偏好键，例如 language 或 target_role"`
	Value any    `json:"value" jsonschema:"description=偏好值，会被 JSON 序列化保存"`
}

// Toolset 负责把应用服务封装成 tRPC-Agent-Go callable tools。
type Toolset struct {
	discovery *appsvc.DiscoveryService
	store     *sqlite.SQLiteStore
	scorer    *scoring.Service
	reporter  *report.Service
}

// NewToolset 创建当前 Agent 需要暴露的工具集合。
func NewToolset(discovery *appsvc.DiscoveryService, store *sqlite.SQLiteStore) *Toolset {
	return &Toolset{
		discovery: discovery,
		store:     store,
		scorer:    scoring.NewService(),
		reporter:  report.NewService(),
	}
}

// Tools 返回 tRPC-Agent-Go 可识别的工具声明和调用实现。
func (t *Toolset) Tools() []tool.Tool {
	return []tool.Tool{
		function.NewFunctionTool(t.searchRepositories,
			function.WithName("search_repositories"),
			function.WithDescription("Search, profile, score, and rank GitHub repositories from a user's learning or recruiting goal."),
		),
		function.NewFunctionTool(t.getRepositoryMetadata,
			function.WithName("get_repository_metadata"),
			function.WithDescription("Read GitHub repository metadata from cache or GitHub; read-only."),
		),
		function.NewFunctionTool(t.getReadme,
			function.WithName("get_readme"),
			function.WithDescription("Read and summarize repository README; read-only and deterministic."),
		),
		function.NewFunctionTool(t.getTree,
			function.WithName("get_tree"),
			function.WithDescription("Read repository tree and summarize architecture directories; read-only."),
		),
		function.NewFunctionTool(t.getDependencyFiles,
			function.WithName("get_dependency_files"),
			function.WithDescription("Detect dependency files from the repository tree; read-only."),
		),
		function.NewFunctionTool(t.classifyIssues,
			function.WithName("classify_issues"),
			function.WithDescription("Classify open issues into docs, bug, feature, test, good-first, and infra buckets; read-only."),
		),
		function.NewFunctionTool(t.summarizePRs,
			function.WithName("summarize_prs"),
			function.WithDescription("Summarize open pull request risk signals; read-only."),
		),
		function.NewFunctionTool(t.scoreRepository,
			function.WithName("score_repository"),
			function.WithDescription("Score one repository profile with deterministic activity, popularity, learning, contribution, and role relevance rules."),
		),
		function.NewFunctionTool(t.generateContributionPlan,
			function.WithName("generate_contribution_plan"),
			function.WithDescription("Generate a deterministic seven-day contribution plan and resume/interview value summary."),
		),
		function.NewFunctionTool(t.generateProjectReport,
			function.WithName("generate_project_report"),
			function.WithDescription("Generate a Markdown recommendation report from a discovery result."),
		),
		function.NewFunctionTool(t.rememberUserPreference,
			function.WithName("remember_user_preference"),
			function.WithDescription("Persist a long-term user preference in the local SQLite store."),
		),
	}
}

func (t *Toolset) searchRepositories(ctx context.Context, input SearchRepositoriesInput) (domain.DiscoveryResult, error) {
	if t.discovery == nil {
		return domain.DiscoveryResult{}, errors.New("discovery service is not configured")
	}
	return t.discovery.DiscoverProjects(ctx, input)
}

func (t *Toolset) getRepositoryMetadata(ctx context.Context, input RepositoryInput) (domain.Repository, error) {
	if t.discovery == nil {
		return domain.Repository{}, errors.New("discovery service is not configured")
	}
	if input.FullName == "" {
		return domain.Repository{}, errors.New("full_name is required")
	}
	return t.discovery.GetRepositoryMetadata(ctx, input.FullName)
}

func (t *Toolset) getReadme(ctx context.Context, input ReadmeInput) (string, error) {
	if t.discovery == nil {
		return "", errors.New("discovery service is not configured")
	}
	if input.Repository.FullName == "" {
		return "", errors.New("repository.full_name is required")
	}
	return t.discovery.GetReadmeSummary(ctx, input.Repository), nil
}

func (t *Toolset) getTree(ctx context.Context, input TreeInput) (TreeResult, error) {
	if t.discovery == nil {
		return TreeResult{}, errors.New("discovery service is not configured")
	}
	if input.Repository.FullName == "" {
		return TreeResult{}, errors.New("repository.full_name is required")
	}
	directorySummary, dependencyFiles, dependencySummary := t.discovery.GetTreeSummary(ctx, input.Repository)
	return TreeResult{DirectorySummary: directorySummary, DependencyFiles: dependencyFiles, DependencySummary: dependencySummary}, nil
}

func (t *Toolset) getDependencyFiles(ctx context.Context, input TreeInput) (TreeResult, error) {
	return t.getTree(ctx, input)
}

func (t *Toolset) classifyIssues(ctx context.Context, input RepositoryInput) (string, error) {
	if t.discovery == nil {
		return "", errors.New("discovery service is not configured")
	}
	if input.FullName == "" {
		return "", errors.New("full_name is required")
	}
	return t.discovery.ClassifyIssues(ctx, input.FullName), nil
}

func (t *Toolset) summarizePRs(ctx context.Context, input RepositoryInput) (string, error) {
	if t.discovery == nil {
		return "", errors.New("discovery service is not configured")
	}
	if input.FullName == "" {
		return "", errors.New("full_name is required")
	}
	return t.discovery.SummarizePullRequests(ctx, input.FullName), nil
}

func (t *Toolset) scoreRepository(_ context.Context, input ScoreRepositoryInput) (domain.Score, error) {
	return t.scorer.Score(input.Profile, input.Intent), nil
}

func (t *Toolset) generateContributionPlan(_ context.Context, input ContributionPlanInput) (ContributionPlanResult, error) {
	if t.discovery == nil {
		return ContributionPlanResult{}, errors.New("discovery service is not configured")
	}
	plan, resumeValue := t.discovery.GenerateContributionPlan(input.Analysis)
	return ContributionPlanResult{ContributionPlan: plan, ResumeValue: resumeValue}, nil
}

func (t *Toolset) generateProjectReport(_ context.Context, input GenerateProjectReportInput) (string, error) {
	return t.reporter.Recommendation(input.Result), nil
}

func (t *Toolset) rememberUserPreference(ctx context.Context, input RememberUserPreferenceInput) (string, error) {
	if t.store == nil {
		return "preference skipped: SQLite store is not configured", nil
	}
	if input.Key == "" {
		return "", errors.New("preference key is required")
	}
	if err := t.store.SaveUserPreference(ctx, input.Key, input.Value); err != nil {
		return "", err
	}
	return "preference saved", nil
}

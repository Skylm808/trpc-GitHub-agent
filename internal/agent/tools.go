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
type SearchRepositoriesInput struct {
	UserInput string `json:"user_input" jsonschema:"description=用户自然语言背景和项目目标"`
	Limit     int    `json:"limit" jsonschema:"description=最多返回的项目数量"`
}

// ScoreRepositoryInput 是 score_repository tool 的输入。
type ScoreRepositoryInput struct {
	Profile domain.RepositoryProfile `json:"profile" jsonschema:"description=仓库画像"`
	Intent  domain.SearchIntent      `json:"intent" jsonschema:"description=用户搜索意图"`
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
		function.NewFunctionTool(t.scoreRepository,
			function.WithName("score_repository"),
			function.WithDescription("Score one repository profile with deterministic activity, popularity, learning, contribution, and role relevance rules."),
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
	return t.discovery.DiscoverProjects(ctx, input.UserInput, input.Limit)
}

func (t *Toolset) scoreRepository(_ context.Context, input ScoreRepositoryInput) (domain.Score, error) {
	return t.scorer.Score(input.Profile, input.Intent), nil
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

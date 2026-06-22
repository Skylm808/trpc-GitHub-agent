package app

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	gh "trpc-GitHub-agent/internal/clients/github"
	"trpc-GitHub-agent/internal/domain"
	"trpc-GitHub-agent/internal/services/queryplanner"
	"trpc-GitHub-agent/internal/services/report"
	"trpc-GitHub-agent/internal/services/scoring"
	"trpc-GitHub-agent/internal/store/sqlite"
)

// DiscoveryService 编排项目发现、仓库画像、评分、报告和缓存写入。
type DiscoveryService struct {
	store    *sqlite.SQLiteStore
	github   *gh.Client
	planner  *queryplanner.Planner
	scorer   *scoring.Service
	reporter *report.Service
}

// NewDiscoveryService 创建项目发现服务；store 可为空，表示只运行无缓存模式。
func NewDiscoveryService(github *gh.Client, store *sqlite.SQLiteStore) *DiscoveryService {
	return &DiscoveryService{
		store:    store,
		github:   github,
		planner:  queryplanner.NewPlanner(),
		scorer:   scoring.NewService(),
		reporter: report.NewService(),
	}
}

// SetStore 在 Wails 启动后注入 SQLite 缓存。
func (s *DiscoveryService) SetStore(store *sqlite.SQLiteStore) {
	s.store = store
}

// DiscoverProjects 根据用户自然语言目标和结构化筛选生成 GitHub 查询，并返回排序后的项目研究结果。
func (s *DiscoveryService) DiscoverProjects(ctx context.Context, request domain.SearchRequest) (domain.DiscoveryResult, error) {
	userInput := request.UserInput
	if strings.TrimSpace(userInput) == "" {
		userInput = "Go backend Agent projects for learning and open source contribution"
	}
	filters := domain.SearchIntent{
		InputLanguage: request.InputLanguage,
		Languages:     request.Languages,
		Topics:        request.Topics,
		TargetRole:    request.TargetRole,
		Difficulty:    request.Difficulty,
		Direction:     request.Direction,
		PushedAfter:   request.PushedAfter,
		MinStars:      request.MinStars,
		MaxStars:      request.MaxStars,
	}
	intent, queries := s.planner.Plan(userInput, request.Limit, filters)
	limit := request.Limit
	if limit <= 0 {
		limit = intent.ProjectSize
	}
	if limit <= 0 {
		limit = 10
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var warnings []string
	repos, usedLive := s.searchLiveRepositories(ctx, queries, limit, &warnings)
	if len(repos) == 0 {
		repos = fallbackRepositories()
		usedLive = false
		warnings = append(warnings, "GitHub live search unavailable or empty; showing deterministic demo fixtures.")
	}

	scored := make([]domain.ScoredRepository, 0, len(repos))
	var repoIDs []int64
	for _, repo := range repos {
		repoID := repo.ID
		if s.store != nil {
			if savedID, err := s.store.UpsertRepository(ctx, repo); err == nil {
				repo.ID = savedID
				repoID = savedID
				repoIDs = append(repoIDs, savedID)
			} else {
				warnings = append(warnings, "SQLite cache write failed for "+repo.FullName+": "+err.Error())
			}
		}
		profile := s.profileRepository(ctx, repo, usedLive, &warnings)
		profile.Repository.ID = repoID
		score := s.scorer.Score(profile, intent)
		if s.store != nil && repoID > 0 {
			_ = s.store.SaveScore(ctx, score)
		}
		scored = append(scored, domain.ScoredRepository{Repository: profile.Repository, Score: score})
	}
	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].Score.TotalScore > scored[j].Score.TotalScore
	})
	if len(scored) > limit {
		scored = scored[:limit]
	}

	result := domain.DiscoveryResult{
		Intent:         intent,
		Queries:        queries,
		Repositories:   scored,
		UsedLiveGitHub: usedLive,
		Warnings:       warnings,
	}
	result.MarkdownReport = s.reporter.Recommendation(result)
	result.AgentTrace = buildDiscoveryTrace(result)
	if s.store != nil {
		_ = s.store.SaveQueryHistory(ctx, intent, queries, repoIDs)
		if len(scored) > 0 {
			_ = s.store.SaveReport(ctx, scored[0].Repository.ID, "recommendation", "GitHub Project Recommendation Report", result.MarkdownReport)
		}
	}
	return result, nil
}

func buildDiscoveryTrace(result domain.DiscoveryResult) []domain.AgentTraceStep {
	source := "GitHub live search"
	if !result.UsedLiveGitHub {
		source = "deterministic fixture fallback"
	}
	return []domain.AgentTraceStep{
		{Phase: "Plan", Tool: "plan_search_queries", Summary: "从用户输入中抽取语言、方向、难度、Star 范围和最近活跃时间，生成 GitHub 查询。"},
		{Phase: "Tool Calls", Tool: "search_repositories", Summary: "执行 " + strconv.Itoa(len(result.Queries)) + " 条 GitHub repository search query；数据来源：" + source + "。"},
		{Phase: "Tool Calls", Tool: "score_repository", Summary: "对 " + strconv.Itoa(len(result.Repositories)) + " 个仓库执行确定性评分，LLM 不参与排序。"},
		{Phase: "Findings", Tool: "rank_repositories", Summary: "按确定性总分生成项目分级列表，用于后续单仓研究。"},
		{Phase: "Report", Tool: "generate_project_report", Summary: "生成 Markdown 研究报告；AI 润色仍与确定性评分分离。"},
	}
}

func (s *DiscoveryService) searchLiveRepositories(ctx context.Context, queries []domain.PlannedQuery, limit int, warnings *[]string) ([]domain.Repository, bool) {
	if s.github == nil {
		return nil, false
	}
	seen := map[string]bool{}
	var repos []domain.Repository
	perQuery := limit
	if len(queries) > 1 {
		perQuery = max(3, limit/len(queries)+1)
	}
	for _, planned := range queries {
		found, err := s.github.SearchRepositories(ctx, planned.Query, perQuery)
		if err != nil {
			*warnings = append(*warnings, "GitHub query failed: "+err.Error())
			continue
		}
		for _, repo := range found {
			if seen[repo.FullName] {
				continue
			}
			seen[repo.FullName] = true
			repos = append(repos, repo)
			if len(repos) >= limit {
				return repos, true
			}
		}
	}
	return repos, len(repos) > 0
}

func (s *DiscoveryService) profileRepository(ctx context.Context, repo domain.Repository, live bool, warnings *[]string) domain.RepositoryProfile {
	profile := domain.RepositoryProfile{
		Repository:       repo,
		HasReadme:        true,
		StructureSummary: "Repository metadata available from GitHub search.",
	}
	if !live || s.github == nil {
		applyFixtureSignals(&profile)
		return profile
	}

	readme, err := s.github.GetReadme(ctx, repo.FullName)
	if err == nil && strings.TrimSpace(readme) != "" {
		profile.HasReadme = true
		profile.ReadmeSummary = summarizeReadme(readme)
	} else if err != nil {
		*warnings = append(*warnings, "README fetch failed for "+repo.FullName+": "+err.Error())
	}

	tree, err := s.github.GetTree(ctx, repo.FullName, repo.DefaultBranch)
	if err == nil {
		paths := make([]string, 0, len(tree))
		for _, item := range tree {
			paths = append(paths, item.Path)
		}
		hasDocs, hasExamples, hasTests, hasContributing, dependencyFiles, dependencySummary, structureSummary := sqlite.AnalyzeTree(paths)
		profile.HasDocs = hasDocs
		profile.HasExamples = hasExamples
		profile.HasTests = hasTests
		profile.HasContributing = hasContributing
		profile.DependencySummary = dependencySummary
		profile.StructureSummary = structureSummary
		_ = dependencyFiles
	} else {
		*warnings = append(*warnings, "Tree fetch failed for "+repo.FullName+": "+err.Error())
	}

	if count, err := s.github.CountIssuesByLabel(ctx, repo.FullName, "good first issue"); err == nil {
		profile.GoodFirstIssueCount = count
	}
	if count, err := s.github.CountIssuesByLabel(ctx, repo.FullName, "help wanted"); err == nil {
		profile.HelpWantedCount = count
	}
	return profile
}

// AnalyzeRepository 返回单仓库深度分析结果，用于详情视图和后续 LLM 总结。
func (s *DiscoveryService) AnalyzeRepository(ctx context.Context, fullName string) (domain.RepositoryAnalysis, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	repo, ok, err := s.lookupRepository(ctx, fullName)
	if err != nil {
		return domain.RepositoryAnalysis{}, err
	}
	if !ok {
		repo = domain.Repository{FullName: fullName, HTMLURL: "https://github.com/" + fullName}
	}
	profile := s.profileRepository(ctx, repo, s.github != nil, &[]string{})
	paths := []string{}
	if s.github != nil {
		if tree, err := s.github.GetTree(ctx, profile.Repository.FullName, profile.Repository.DefaultBranch); err == nil {
			for _, item := range tree {
				paths = append(paths, item.Path)
			}
		}
	}
	_, _, _, _, dependencyFiles, dependencySummary, structureSummary := sqlite.AnalyzeTree(paths)
	issueSummary := s.analyzeIssues(ctx, profile.Repository.FullName)
	prSummary := s.analyzePullRequests(ctx, profile.Repository.FullName)
	analysis := domain.RepositoryAnalysis{
		Repository:        profile.Repository,
		Profile:           profile,
		Positioning:       summarizePositioning(profile.Repository),
		Architecture:      summarizeArchitecture(structureSummary, profile.StructureSummary),
		LearningModules:   suggestLearningModules(profile, dependencyFiles),
		ContributionTypes: suggestContributionTypes(issueSummary, profile),
		DocsSummary:       summarizeSignal(profile.HasDocs, "检测到 docs 目录或文档文件"),
		ExamplesSummary:   summarizeSignal(profile.HasExamples, "检测到 examples 示例目录"),
		TestsSummary:      summarizeSignal(profile.HasTests, "检测到测试目录或测试文件"),
		DependencyFiles:   dependencyFiles,
		DirectorySummary:  structureSummary,
		IssueSummary:      issueSummary,
		PRSummary:         prSummary,
	}
	if analysis.DirectorySummary == "" {
		analysis.DirectorySummary = profile.StructureSummary
	}
	if analysis.DependencyFiles == nil {
		analysis.DependencyFiles = []string{}
	}
	if dependencySummary != "" && analysis.Profile.DependencySummary == "" {
		analysis.Profile.DependencySummary = dependencySummary
	}
	if strings.TrimSpace(analysis.Profile.ReadmeSummary) == "" {
		analysis.Profile.ReadmeSummary = fallbackReadmeSummary(analysis.Repository)
	}
	analysis.ContributionPlan = generateContributionPlan(analysis)
	analysis.ResumeValue = summarizeResumeValue(analysis.Repository, analysis.Profile)
	analysis.AgentTrace = buildAgentTrace(analysis)
	return analysis, nil
}

// GetRepositoryMetadata 返回仓库基础元数据，只读且优先使用本地缓存。
func (s *DiscoveryService) GetRepositoryMetadata(ctx context.Context, fullName string) (domain.Repository, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	repo, ok, err := s.lookupRepository(ctx, fullName)
	if err != nil {
		return domain.Repository{}, err
	}
	if ok {
		return repo, nil
	}
	return domain.Repository{FullName: fullName, HTMLURL: "https://github.com/" + fullName}, nil
}

// GetReadmeSummary 返回 README 摘要；GitHub 不可用时退化为仓库描述。
func (s *DiscoveryService) GetReadmeSummary(ctx context.Context, repo domain.Repository) string {
	if ctx == nil {
		ctx = context.Background()
	}
	if s.github != nil {
		if readme, err := s.github.GetReadme(ctx, repo.FullName); err == nil && strings.TrimSpace(readme) != "" {
			return summarizeReadme(readme)
		}
	}
	return fallbackReadmeSummary(repo)
}

// GetTreeSummary 返回目录树摘要和依赖文件列表。
func (s *DiscoveryService) GetTreeSummary(ctx context.Context, repo domain.Repository) (string, []string, string) {
	if ctx == nil {
		ctx = context.Background()
	}
	var paths []string
	if s.github != nil {
		if tree, err := s.github.GetTree(ctx, repo.FullName, repo.DefaultBranch); err == nil {
			for _, item := range tree {
				paths = append(paths, item.Path)
			}
		}
	}
	_, _, _, _, dependencyFiles, dependencySummary, structureSummary := sqlite.AnalyzeTree(paths)
	if structureSummary == "" {
		structureSummary = "目录树暂不可用；建议配置 GitHub token 后重新分析，以减少 API 限流或网络失败。"
	}
	if dependencyFiles == nil {
		dependencyFiles = []string{}
	}
	return structureSummary, dependencyFiles, dependencySummary
}

// ClassifyIssues 按固定标签集合分类开放 Issue，保持只读。
func (s *DiscoveryService) ClassifyIssues(ctx context.Context, fullName string) string {
	if ctx == nil {
		ctx = context.Background()
	}
	return s.analyzeIssues(ctx, fullName)
}

// SummarizePullRequests 总结开放 PR 风险，保持只读。
func (s *DiscoveryService) SummarizePullRequests(ctx context.Context, fullName string) string {
	if ctx == nil {
		ctx = context.Background()
	}
	return s.analyzePullRequests(ctx, fullName)
}

// GenerateContributionPlan 输出确定性的 7 天贡献计划和简历价值。
func (s *DiscoveryService) GenerateContributionPlan(analysis domain.RepositoryAnalysis) (string, string) {
	return generateContributionPlan(analysis), summarizeResumeValue(analysis.Repository, analysis.Profile)
}

func (s *DiscoveryService) analyzeIssues(ctx context.Context, fullName string) string {
	if s.github == nil {
		return summarizeIssueSummary(domain.RepositoryProfile{})
	}
	groups := map[string]int{}
	labels := []string{"docs", "bug", "feature", "test", "good first issue", "infra"}
	for _, label := range labels {
		issues, err := s.github.SearchIssues(ctx, fullName, "is:open label:\""+label+"\"", 10)
		if err != nil {
			continue
		}
		if len(issues) > 0 {
			groups[label] = len(issues)
		}
	}
	if len(groups) == 0 {
		return "未从 GitHub 搜索结果中识别到可分类的开放 Issue 标签"
	}
	var parts []string
	for _, label := range []string{"docs", "bug", "feature", "test", "good first issue", "infra"} {
		if count := groups[label]; count > 0 {
			parts = append(parts, issueLabelName(label)+":"+strconv.Itoa(count))
		}
	}
	return "识别到 Issue 分类 -> " + strings.Join(parts, ", ")
}

func (s *DiscoveryService) analyzePullRequests(ctx context.Context, fullName string) string {
	if s.github == nil {
		return summarizePRSummary(domain.RepositoryProfile{})
	}
	prs, err := s.github.SearchPullRequests(ctx, fullName, "is:open", 10)
	if err != nil {
		return "PR 搜索暂不可用"
	}
	if len(prs) == 0 {
		return "未发现开放 PR"
	}
	var riskSignals []string
	for _, pr := range prs {
		title := strings.ToLower(pr.Title)
		switch {
		case strings.Contains(title, "refactor"), strings.Contains(title, "breaking"):
			riskSignals = append(riskSignals, "高风险变更")
		case strings.Contains(title, "test"), strings.Contains(title, "docs"):
			riskSignals = append(riskSignals, "低风险")
		default:
			riskSignals = append(riskSignals, "中等风险")
		}
	}
	return "开放 PR 数量：" + strconv.Itoa(len(prs)) + "，风险信号：" + strings.Join(riskSignals, ", ")
}

func (s *DiscoveryService) lookupRepository(ctx context.Context, fullName string) (domain.Repository, bool, error) {
	if s.store == nil {
		return domain.Repository{}, false, nil
	}
	repo, ok, err := s.store.GetRepositoryByFullName(ctx, fullName)
	if err != nil {
		return domain.Repository{}, false, err
	}
	return repo, ok, nil
}

func summarizeSignal(enabled bool, fallback string) string {
	if enabled {
		return fallback
	}
	return "未检测到该信号"
}

func summarizeIssueSummary(profile domain.RepositoryProfile) string {
	switch {
	case profile.GoodFirstIssueCount > 0 && profile.HelpWantedCount > 0:
		return "检测到 good-first-issue 和 help-wanted，适合作为贡献入口"
	case profile.GoodFirstIssueCount > 0:
		return "检测到 good-first-issue，适合新贡献者切入"
	case profile.HelpWantedCount > 0:
		return "检测到 help-wanted，维护者有明确协作需求"
	default:
		return "未检测到明确的贡献导向 Issue 标签"
	}
}

func summarizePRSummary(profile domain.RepositoryProfile) string {
	if profile.HasTests && profile.HasContributing {
		return "仓库同时具备测试和贡献指南，PR 风险中等且较可控"
	}
	if profile.HasTests {
		return "仓库有测试，但贡献指南较有限，提交 PR 前需要多读现有规范"
	}
	return "仅凭当前元数据较难判断 PR 风险，建议进一步阅读近期 PR 和维护者反馈"
}

func summarizePositioning(repo domain.Repository) string {
	parts := []string{}
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
		description = "适合从 README、目录树和 Issue 信号继续判断项目方向。"
	}
	return repo.FullName + " 定位为 " + strings.Join(parts, " / ") + "；" + description
}

func summarizeArchitecture(directorySummary, profileSummary string) string {
	if strings.TrimSpace(directorySummary) != "" {
		return directorySummary
	}
	if strings.TrimSpace(profileSummary) != "" {
		return profileSummary
	}
	return "目录结构暂不可用；建议配置 GitHub token 后重新分析 README、tree、docs、examples 和 tests。"
}

func suggestLearningModules(profile domain.RepositoryProfile, dependencyFiles []string) []string {
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

func suggestContributionTypes(issueSummary string, profile domain.RepositoryProfile) []string {
	lower := strings.ToLower(issueSummary)
	candidates := []struct {
		token string
		label string
	}{
		{"文档", "docs"},
		{"docs", "docs"},
		{"缺陷", "bug"},
		{"bug", "bug"},
		{"功能", "feature"},
		{"feature", "feature"},
		{"测试", "test"},
		{"test", "test"},
		{"good-first", "good-first"},
		{"infra", "infra"},
	}
	seen := map[string]bool{}
	var result []string
	for _, candidate := range candidates {
		if strings.Contains(lower, strings.ToLower(candidate.token)) && !seen[candidate.label] {
			seen[candidate.label] = true
			result = append(result, candidate.label)
		}
	}
	if profile.GoodFirstIssueCount > 0 && !seen["good-first"] {
		result = append(result, "good-first")
		seen["good-first"] = true
	}
	if profile.HasDocs && !seen["docs"] {
		result = append(result, "docs")
	}
	if profile.HasTests && !seen["test"] {
		result = append(result, "test")
	}
	if len(result) == 0 {
		result = append(result, "docs", "test", "good-first")
	}
	return result
}

func summarizeReadme(readme string) string {
	readme = strings.TrimSpace(readme)
	if len(readme) <= 280 {
		return readme
	}
	return readme[:280] + "..."
}

func fallbackReadmeSummary(repo domain.Repository) string {
	if strings.TrimSpace(repo.Description) != "" {
		return "README 摘要暂不可用，先使用仓库描述：" + strings.TrimSpace(repo.Description)
	}
	return "README 摘要暂不可用。建议配置 GitHub token 后重新分析，以减少 API 限流导致的读取失败。"
}

func generateContributionPlan(analysis domain.RepositoryAnalysis) string {
	return strings.Join([]string{
		"Day 1：阅读 README、贡献指南和目录树，确认项目定位与核心模块。",
		"Day 2：运行 examples 或 tests，记录环境问题和可复现步骤。",
		"Day 3：从 docs/test/good-first 类型 Issue 中挑选范围最小的一项。",
		"Day 4：阅读相关模块代码，补充最小测试或文档草稿。",
		"Day 5：提交小范围 PR，说明变更动机、影响范围和验证结果。",
		"Day 6：根据维护者反馈迭代，避免扩大 PR 范围。",
		"Day 7：沉淀贡献复盘，用于简历项目经历或面试讲述。",
	}, "\n")
}

func summarizeResumeValue(repo domain.Repository, profile domain.RepositoryProfile) string {
	signals := []string{}
	if profile.HasDocs {
		signals = append(signals, "文档")
	}
	if profile.HasExamples {
		signals = append(signals, "示例")
	}
	if profile.HasTests {
		signals = append(signals, "测试")
	}
	if profile.HasContributing {
		signals = append(signals, "贡献指南")
	}
	if len(signals) == 0 {
		signals = append(signals, "源码阅读")
	}
	return repo.FullName + " 适合沉淀为开源贡献/源码阅读经历，可围绕 " + strings.Join(signals, "、") + " 展开 STAR 叙事。"
}

func buildAgentTrace(analysis domain.RepositoryAnalysis) []domain.AgentTraceStep {
	return []domain.AgentTraceStep{
		{Phase: "Plan", Tool: "AnalyzeRepository", Summary: "制定只读研究流程：元数据 -> README -> 目录树 -> Issue/PR -> 贡献计划。"},
		{Phase: "Tool Calls", Tool: "get_repository_metadata", Summary: "读取仓库基础信息、语言、stars、forks、open issues 和默认分支。"},
		{Phase: "Tool Calls", Tool: "get_readme", Summary: analysis.Profile.ReadmeSummary},
		{Phase: "Tool Calls", Tool: "get_tree / get_dependency_files", Summary: analysis.DirectorySummary},
		{Phase: "Tool Calls", Tool: "classify_issues", Summary: analysis.IssueSummary},
		{Phase: "Tool Calls", Tool: "summarize_prs", Summary: analysis.PRSummary},
		{Phase: "Findings", Tool: "generate_contribution_plan", Summary: "生成 7 天贡献路线和简历价值总结；LLM 输出区域仍与确定性评分分离。"},
	}
}

func issueLabelName(label string) string {
	switch label {
	case "docs":
		return "文档"
	case "bug":
		return "缺陷"
	case "feature":
		return "功能"
	case "test":
		return "测试"
	case "good first issue":
		return "good-first"
	case "infra":
		return "基础设施"
	default:
		return label
	}
}

func applyFixtureSignals(profile *domain.RepositoryProfile) {
	text := strings.ToLower(profile.Repository.FullName + " " + profile.Repository.Description)
	profile.HasReadme = true
	profile.HasDocs = true
	profile.HasExamples = strings.Contains(text, "agent") || strings.Contains(text, "rag")
	profile.HasTests = true
	profile.HasContributing = profile.Repository.Stars < 10000
	profile.GoodFirstIssueCount = 3
	profile.HelpWantedCount = 2
	profile.DependencySummary = "Dependency files: go.mod"
	profile.StructureSummary = "Demo profile includes README, docs, examples, tests, and dependency signals."
	profile.ReadmeSummary = fallbackReadmeSummary(profile.Repository)
}

func fallbackRepositories() []domain.Repository {
	now := time.Now().UTC()
	return []domain.Repository{
		fixtureRepo(1, "trpc-group/trpc-agent-go", "Agent framework for building LLM applications in Go with tools, memory, graph, and evaluation.", "Go", []string{"agent", "llm-agent", "framework"}, 2500, 280, now.AddDate(0, 0, -10)),
		fixtureRepo(2, "mark3labs/mcp-go", "Go SDK for the Model Context Protocol.", "Go", []string{"mcp", "model-context-protocol"}, 4200, 410, now.AddDate(0, 0, -8)),
		fixtureRepo(3, "ollama/ollama", "Run large language models locally.", "Go", []string{"llm", "runtime"}, 180000, 14000, now.AddDate(0, 0, -2)),
		fixtureRepo(4, "langchain-ai/langchaingo", "LangChain for Go, with chains, agents, tools, and retrieval integrations.", "Go", []string{"rag", "agent", "retrieval"}, 6500, 900, now.AddDate(0, 0, -20)),
		fixtureRepo(5, "kubernetes-sigs/kubebuilder", "SDK for building Kubernetes APIs using CRDs.", "Go", []string{"backend", "framework"}, 7500, 1500, now.AddDate(0, -2, 0)),
	}
}

func fixtureRepo(id int64, fullName, description, language string, topics []string, stars, forks int, pushedAt time.Time) domain.Repository {
	parts := strings.Split(fullName, "/")
	return domain.Repository{
		GitHubID:        id,
		FullName:        fullName,
		Owner:           parts[0],
		Name:            parts[1],
		Description:     description,
		HTMLURL:         "https://github.com/" + fullName,
		CloneURL:        "https://github.com/" + fullName + ".git",
		Language:        language,
		Topics:          topics,
		Stars:           stars,
		Forks:           forks,
		Watchers:        stars,
		OpenIssuesCount: 30,
		DefaultBranch:   "main",
		PushedAt:        pushedAt,
		CreatedAt:       pushedAt.AddDate(-2, 0, 0),
		UpdatedAt:       pushedAt,
		FetchedAt:       time.Now().UTC(),
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

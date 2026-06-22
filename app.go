package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	researchagent "trpc-GitHub-agent/internal/agent"
	appsvc "trpc-GitHub-agent/internal/app"
	gh "trpc-GitHub-agent/internal/clients/github"
	"trpc-GitHub-agent/internal/clients/llmchat"
	"trpc-GitHub-agent/internal/config"
	"trpc-GitHub-agent/internal/domain"
	"trpc-GitHub-agent/internal/store/sqlite"
)

type App struct {
	ctx       context.Context
	store     *sqlite.SQLiteStore
	discovery *appsvc.DiscoveryService
	agent     *researchagent.Runner
	storePath string
}

// NewApp 创建 Wails 根对象，并初始化可在无 SQLite 时运行的发现服务。
func NewApp() *App {
	discovery := appsvc.NewDiscoveryService(newGitHubClientFromConfig(), nil)
	return &App{discovery: discovery, agent: researchagent.NewRunner(discovery, nil)}
}

// startup 在 Wails 启动时打开本地 SQLite 缓存。
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	sqliteStore, err := sqlite.Open("")
	if err == nil {
		a.store = sqliteStore
		a.storePath, _ = sqlite.DefaultPath()
		a.discovery.SetStore(sqliteStore)
		a.agent = researchagent.NewRunner(a.discovery, sqliteStore)
	}
}

// shutdown 关闭本地 SQLite 连接。
func (a *App) shutdown(ctx context.Context) {
	if a.store != nil {
		_ = a.store.Close()
	}
}

// DiscoverProjects 暴露给前端：根据用户背景和筛选条件发现、评分并生成项目报告。
func (a *App) DiscoverProjects(request domain.SearchRequest) (domain.DiscoveryResult, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.agent.DiscoverProjectsWithRequest(ctx, request)
}

// StorePath 返回当前 SQLite 缓存文件路径，便于用户定位本地缓存。
func (a *App) StorePath() string {
	return a.storePath
}

// SettingsStatus 返回 GitHub 与 LLM provider key 的本地配置状态。
func (a *App) SettingsStatus() config.SettingsStatus {
	return config.LoadSettingsStatus()
}

// Settings 返回可编辑的本地配置、密钥状态和文件路径。
func (a *App) Settings() config.SettingsBundle {
	return config.LoadSettingsBundle()
}

// SaveSettings 保存本地 YAML 配置和密钥文件。
func (a *App) SaveSettings(update config.SettingsUpdate) (config.SettingsBundle, error) {
	bundle, err := config.SaveSettings(update)
	if err != nil {
		return bundle, err
	}
	a.refreshServices()
	return bundle, nil
}

// AnalyzeRepository 暴露给前端：分析单个仓库的 README、目录树、Issues 和 PR 信号。
func (a *App) AnalyzeRepository(fullName string) (domain.RepositoryAnalysis, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	analysis, err := a.agent.AnalyzeRepository(ctx, fullName)
	if err != nil {
		runnerErr := err
		analysis, err = a.discovery.AnalyzeRepository(ctx, fullName)
		if err != nil {
			return domain.RepositoryAnalysis{}, err
		}
		analysis.AgentTrace = append(analysis.AgentTrace, domain.AgentTraceStep{Phase: "Findings", Tool: "agent_runner_fallback", Summary: "Runner 工具链失败，已回退到 DiscoveryService：" + runnerErr.Error()})
	}
	analysis = a.enhanceRepositoryAnalysis(ctx, analysis)
	a.saveResearchSession(ctx, &analysis)
	return analysis, nil
}

// ListResearchSessions 返回最近的研究会话，供前端回看历史分析。
func (a *App) ListResearchSessions(limit int) ([]domain.ResearchSession, error) {
	if a.store == nil {
		return []domain.ResearchSession{}, nil
	}
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return a.store.ListResearchSessions(ctx, limit)
}

// GetResearchSession 读取一次历史研究会话。
func (a *App) GetResearchSession(id int64) (domain.ResearchSession, error) {
	if a.store == nil {
		return domain.ResearchSession{}, fmt.Errorf("SQLite store 未启用，无法读取研究会话")
	}
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	session, ok, err := a.store.GetResearchSession(ctx, id)
	if err != nil {
		return domain.ResearchSession{}, err
	}
	if !ok {
		return domain.ResearchSession{}, fmt.Errorf("研究会话不存在：%d", id)
	}
	return session, nil
}

// TestGitHubConnection 做本地配置完整性检测，避免在 UI 中误把 token 写入 YAML。
func (a *App) TestGitHubConnection() config.ConnectionCheck {
	return config.CheckGitHubConnection()
}

// TestLLMConnection 检查指定 LLM provider 的 base URL、启用状态和 token 配置。
func (a *App) TestLLMConnection(provider string) config.ConnectionCheck {
	return config.CheckLLMConnection(provider)
}

// TestLLMConnectionDraft 使用前端当前表单值进行实时检测，不保存到本地配置。
func (a *App) TestLLMConnectionDraft(request config.ProviderConnectionRequest) config.ConnectionCheck {
	return config.CheckLLMConnectionRequest(request)
}

// AskRepositoryQuestion 使用当前 LLM provider 对单仓库分析结果进行只读问答。
func (a *App) AskRepositoryQuestion(request domain.RepositoryQuestionRequest) (domain.RepositoryQuestionResponse, error) {
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	analysis, err := a.discovery.AnalyzeRepository(ctx, request.FullName)
	if err != nil {
		return domain.RepositoryQuestionResponse{}, err
	}
	runtimeProvider, ok := config.ActiveRuntimeProvider()
	if !ok {
		return domain.RepositoryQuestionResponse{}, fmt.Errorf("LLM provider 未配置或 token/base URL 缺失，请先在 Settings 中配置并检测")
	}
	client := llmchat.Client{
		Provider: runtimeProvider.Provider.Name,
		Model:    runtimeProvider.Provider.Model,
		BaseURL:  runtimeProvider.Provider.BaseURL,
		Token:    runtimeProvider.Token,
	}
	messages := []llmchat.Message{
		{Role: "system", Content: "你是 GitHub 开源贡献研究 Agent。只根据给定仓库上下文和历史问答回答，保持中文，区分确定性评分和 AI 生成建议，不要建议任何 GitHub 写操作。"},
		{Role: "user", Content: repositoryQuestionContextPrompt(analysis)},
		{Role: "assistant", Content: "已读取仓库研究上下文。后续回答会基于该上下文、历史问答和当前问题。"},
	}
	messages = append(messages, repositoryHistoryMessages(request.History)...)
	messages = append(messages, llmchat.Message{Role: "user", Content: "当前问题：" + request.Question + "\n\n请输出：结论、理由、建议下一步。"})
	answer, err := client.Chat(ctx, messages)
	if err != nil {
		return domain.RepositoryQuestionResponse{}, err
	}
	return domain.RepositoryQuestionResponse{
		FullName:    request.FullName,
		Question:    request.Question,
		Answer:      answer,
		Provider:    runtimeProvider.Provider.Name,
		Model:       runtimeProvider.Provider.Model,
		AIGenerated: true,
	}, nil
}

func (a *App) refreshServices() {
	discovery := appsvc.NewDiscoveryService(newGitHubClientFromConfig(), a.store)
	a.discovery = discovery
	a.agent = researchagent.NewRunner(discovery, a.store)
}

func repositoryQuestionPrompt(analysis domain.RepositoryAnalysis, question string) string {
	return repositoryQuestionContextPrompt(analysis) + "\n\n用户问题：" + question + "\n\n请输出：结论、理由、建议下一步。"
}

func repositoryQuestionContextPrompt(analysis domain.RepositoryAnalysis) string {
	return "仓库：" + analysis.Repository.FullName + "\n" +
		"描述：" + analysis.Repository.Description + "\n" +
		"项目定位：" + analysis.Positioning + "\n" +
		"架构理解：" + analysis.Architecture + "\n" +
		"适合学习的模块：" + strings.Join(analysis.LearningModules, ", ") + "\n" +
		"适合贡献的 Issue 类型：" + strings.Join(analysis.ContributionTypes, ", ") + "\n" +
		"README 摘要：" + analysis.Profile.ReadmeSummary + "\n" +
		"目录结构：" + analysis.DirectorySummary + "\n" +
		"依赖文件：" + strings.Join(analysis.DependencyFiles, ", ") + "\n" +
		"Issue 分类：" + analysis.IssueSummary + "\n" +
		"PR 风险：" + analysis.PRSummary + "\n" +
		"贡献计划：" + analysis.ContributionPlan
}

func repositoryHistoryMessages(history []domain.RepositoryQuestionTurn) []llmchat.Message {
	if len(history) > 6 {
		history = history[len(history)-6:]
	}
	messages := make([]llmchat.Message, 0, len(history)*2)
	for _, turn := range history {
		question := strings.TrimSpace(turn.Question)
		answer := strings.TrimSpace(turn.Answer)
		if question == "" || answer == "" {
			continue
		}
		messages = append(messages,
			llmchat.Message{Role: "user", Content: "历史问题：" + question},
			llmchat.Message{Role: "assistant", Content: answer},
		)
	}
	return messages
}

func (a *App) enhanceRepositoryAnalysis(ctx context.Context, analysis domain.RepositoryAnalysis) domain.RepositoryAnalysis {
	runtimeProvider, ok := config.ActiveRuntimeProvider()
	if !ok {
		analysis.LLMInsight.GenerationWarning = "LLM provider 未配置，当前仅展示确定性分析。"
		return analysis
	}
	client := llmchat.Client{
		Provider: runtimeProvider.Provider.Name,
		Model:    runtimeProvider.Provider.Model,
		BaseURL:  runtimeProvider.Provider.BaseURL,
		Token:    runtimeProvider.Token,
	}
	answer, err := client.Chat(ctx, []llmchat.Message{
		{Role: "system", Content: "你是 GitHub 开源贡献研究 Agent。必须只输出 JSON，不要 Markdown。所有内容用中文。不要改变确定性评分，不要建议任何 GitHub 写操作。"},
		{Role: "user", Content: repositoryInsightPrompt(analysis)},
	})
	if err != nil {
		analysis.LLMInsight.GenerationWarning = "LLM 增强失败：" + err.Error()
		analysis.AgentTrace = append(analysis.AgentTrace, domain.AgentTraceStep{Phase: "Findings", Tool: "llm_enhance_repository", Summary: analysis.LLMInsight.GenerationWarning})
		return analysis
	}
	var insight domain.LLMInsight
	if err := json.Unmarshal([]byte(cleanJSONAnswer(answer)), &insight); err != nil {
		insight = domain.LLMInsight{Recommendation: answer}
	}
	insight.Provider = runtimeProvider.Provider.Name
	insight.Model = runtimeProvider.Provider.Model
	insight.AIGenerated = true
	analysis.LLMInsight = insight
	analysis.AgentTrace = append(analysis.AgentTrace, domain.AgentTraceStep{Phase: "Report", Tool: "llm_enhance_repository", Summary: "AI generated：已生成 README 摘要、Issue 解释、PR 风险和贡献路线补充。"})
	return analysis
}

func (a *App) saveResearchSession(ctx context.Context, analysis *domain.RepositoryAnalysis) {
	if a.store == nil || analysis == nil {
		return
	}
	id, err := a.store.SaveResearchSession(ctx, *analysis)
	if err != nil {
		analysis.AgentTrace = append(analysis.AgentTrace, domain.AgentTraceStep{Phase: "Findings", Tool: "save_research_session", Summary: "研究会话保存失败：" + err.Error()})
		return
	}
	analysis.AgentTrace = append(analysis.AgentTrace, domain.AgentTraceStep{Phase: "Report", Tool: "save_research_session", Summary: fmt.Sprintf("研究会话已保存，可在历史中回看：#%d", id)})
}

func cleanJSONAnswer(answer string) string {
	answer = strings.TrimSpace(answer)
	answer = strings.TrimPrefix(answer, "```json")
	answer = strings.TrimPrefix(answer, "```")
	answer = strings.TrimSuffix(answer, "```")
	return strings.TrimSpace(answer)
}

func repositoryInsightPrompt(analysis domain.RepositoryAnalysis) string {
	return `请基于以下确定性仓库分析，生成 AI 补充洞察。返回严格 JSON：
{
  "readme_summary": "...",
  "issue_explanation": "...",
  "pr_risk_summary": "...",
  "contribution_plan": "...",
  "recommendation": "..."
}

仓库：` + analysis.Repository.FullName + `
描述：` + analysis.Repository.Description + `
项目定位：` + analysis.Positioning + `
架构理解：` + analysis.Architecture + `
适合学习的模块：` + strings.Join(analysis.LearningModules, ", ") + `
适合贡献的 Issue 类型：` + strings.Join(analysis.ContributionTypes, ", ") + `
README 摘要：` + analysis.Profile.ReadmeSummary + `
目录结构：` + analysis.DirectorySummary + `
依赖文件：` + strings.Join(analysis.DependencyFiles, ", ") + `
docs/examples/tests：` + analysis.DocsSummary + ` / ` + analysis.ExamplesSummary + ` / ` + analysis.TestsSummary + `
Issue 分类：` + analysis.IssueSummary + `
PR 风险：` + analysis.PRSummary + `
确定性贡献计划：` + analysis.ContributionPlan + `
简历价值：` + analysis.ResumeValue
}

func newGitHubClientFromConfig() *gh.Client {
	return gh.NewClient(
		gh.WithBaseURL(config.ResolvedGitHubBaseURL()),
		gh.WithToken(config.ResolvedGitHubToken()),
	)
}

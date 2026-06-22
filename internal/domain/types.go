package domain

import "time"

// SearchIntent 表示从用户自然语言目标中抽取出的结构化搜索意图。
type SearchIntent struct {
	UserInput     string   `json:"user_input"`
	InputLanguage string   `json:"input_language"`
	Languages     []string `json:"languages"`
	Topics        []string `json:"topics"`
	TargetRole    string   `json:"target_role"`
	Goals         []string `json:"goals"`
	Difficulty    string   `json:"difficulty"`
	Direction     string   `json:"direction"`
	PushedAfter   string   `json:"pushed_after"`
	ProjectSize   int      `json:"project_size"`
	MinStars      int      `json:"min_stars"`
	MaxStars      int      `json:"max_stars"`
}

// SearchRequest 表示前端传给发现流程的结构化请求。
type SearchRequest struct {
	UserInput     string   `json:"user_input"`
	Limit         int      `json:"limit"`
	InputLanguage string   `json:"input_language"`
	Languages     []string `json:"languages"`
	Topics        []string `json:"topics"`
	TargetRole    string   `json:"target_role"`
	Difficulty    string   `json:"difficulty"`
	Direction     string   `json:"direction"`
	PushedAfter   string   `json:"pushed_after"`
	MinStars      int      `json:"min_stars"`
	MaxStars      int      `json:"max_stars"`
}

// PlannedQuery 表示系统生成的一条 GitHub Search query 及其生成理由。
type PlannedQuery struct {
	Query       string `json:"query"`
	Reason      string `json:"reason"`
	Description string `json:"description"`
}

// Repository 保存 GitHub 仓库基础元数据和本地缓存 ID。
type Repository struct {
	ID              int64     `json:"id"`
	GitHubID        int64     `json:"github_id"`
	FullName        string    `json:"full_name"`
	Owner           string    `json:"owner"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	HTMLURL         string    `json:"html_url"`
	CloneURL        string    `json:"clone_url"`
	Language        string    `json:"language"`
	Topics          []string  `json:"topics"`
	Stars           int       `json:"stars"`
	Forks           int       `json:"forks"`
	Watchers        int       `json:"watchers"`
	OpenIssuesCount int       `json:"open_issues_count"`
	DefaultBranch   string    `json:"default_branch"`
	Archived        bool      `json:"archived"`
	Disabled        bool      `json:"disabled"`
	PushedAt        time.Time `json:"pushed_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	FetchedAt       time.Time `json:"fetched_at"`
	RawJSON         string    `json:"-"`
}

// RepoSnapshot 保存 README、目录树和依赖文件等仓库快照信息。
type RepoSnapshot struct {
	RepositoryID        int64     `json:"repository_id"`
	ReadmeText          string    `json:"readme_text"`
	TreeJSON            string    `json:"tree_json"`
	DependencyFilesJSON string    `json:"dependency_files_json"`
	DocsSummary         string    `json:"docs_summary"`
	ExamplesSummary     string    `json:"examples_summary"`
	TestsSummary        string    `json:"tests_summary"`
	FetchedAt           time.Time `json:"fetched_at"`
}

// RepositoryProfile 表示仓库分析后的学习和贡献信号。
type RepositoryProfile struct {
	Repository          Repository `json:"repository"`
	ReadmeSummary       string     `json:"readme_summary"`
	StructureSummary    string     `json:"structure_summary"`
	DependencySummary   string     `json:"dependency_summary"`
	HasReadme           bool       `json:"has_readme"`
	HasDocs             bool       `json:"has_docs"`
	HasExamples         bool       `json:"has_examples"`
	HasTests            bool       `json:"has_tests"`
	HasContributing     bool       `json:"has_contributing"`
	GoodFirstIssueCount int        `json:"good_first_issue_count"`
	HelpWantedCount     int        `json:"help_wanted_count"`
}

// RepositoryAnalysis 聚合单仓库深度分析结果，供前端详情和后续 LLM 总结使用。
type RepositoryAnalysis struct {
	Repository        Repository        `json:"repository"`
	Profile           RepositoryProfile `json:"profile"`
	Positioning       string            `json:"positioning"`
	Architecture      string            `json:"architecture"`
	LearningModules   []string          `json:"learning_modules"`
	ContributionTypes []string          `json:"contribution_types"`
	IssueSummary      string            `json:"issue_summary"`
	PRSummary         string            `json:"pr_summary"`
	DocsSummary       string            `json:"docs_summary"`
	ExamplesSummary   string            `json:"examples_summary"`
	TestsSummary      string            `json:"tests_summary"`
	DependencyFiles   []string          `json:"dependency_files"`
	DirectorySummary  string            `json:"directory_summary"`
	ContributionPlan  string            `json:"contribution_plan"`
	ResumeValue       string            `json:"resume_value"`
	AgentTrace        []AgentTraceStep  `json:"agent_trace"`
	LLMInsight        LLMInsight        `json:"llm_insight"`
}

// ResearchSession 是一次单仓库研究 Agent 运行的持久化快照。
type ResearchSession struct {
	ID             int64              `json:"id"`
	Repository     string             `json:"repository"`
	Title          string             `json:"title"`
	Analysis       RepositoryAnalysis `json:"analysis"`
	CreatedAt      time.Time          `json:"created_at"`
	Provider       string             `json:"provider"`
	Model          string             `json:"model"`
	AIGenerated    bool               `json:"ai_generated"`
	TraceStepCount int                `json:"trace_step_count"`
}

// LLMInsight 保存只读研究 Agent 的 AI 生成补充内容，不参与确定性评分。
type LLMInsight struct {
	ReadmeSummary     string `json:"readme_summary"`
	IssueExplanation  string `json:"issue_explanation"`
	PRRiskSummary     string `json:"pr_risk_summary"`
	ContributionPlan  string `json:"contribution_plan"`
	Recommendation    string `json:"recommendation"`
	Provider          string `json:"provider"`
	Model             string `json:"model"`
	AIGenerated       bool   `json:"ai_generated"`
	GenerationWarning string `json:"generation_warning"`
}

// AgentTraceStep 描述一次只读研究 Agent 的计划、工具调用和发现。
type AgentTraceStep struct {
	Phase   string `json:"phase"`
	Tool    string `json:"tool"`
	Summary string `json:"summary"`
}

// RepositoryQuestionRequest 表示单仓库研究问答请求。
type RepositoryQuestionRequest struct {
	FullName string `json:"full_name"`
	Question string `json:"question"`
}

// RepositoryQuestionResponse 表示 LLM 生成的单仓库研究问答结果。
type RepositoryQuestionResponse struct {
	FullName    string `json:"full_name"`
	Question    string `json:"question"`
	Answer      string `json:"answer"`
	Provider    string `json:"provider"`
	Model       string `json:"model"`
	AIGenerated bool   `json:"ai_generated"`
}

// Score 保存仓库评分的五个维度、总分和解释信息。
type Score struct {
	RepositoryID                  int64             `json:"repository_id"`
	ActivityScore                 int               `json:"activity_score"`
	PopularityScore               int               `json:"popularity_score"`
	LearningValueScore            int               `json:"learning_value_score"`
	ContributionFriendlinessScore int               `json:"contribution_friendliness_score"`
	RoleRelevanceScore            int               `json:"role_relevance_score"`
	TotalScore                    int               `json:"total_score"`
	InfluenceLevel                string            `json:"influence_level"`
	BeginnerFriendliness          string            `json:"beginner_friendliness"`
	Difficulty                    string            `json:"difficulty"`
	RecommendationReason          string            `json:"recommendation_reason"`
	Explanation                   map[string]string `json:"explanation"`
	ScoredAt                      time.Time         `json:"scored_at"`
}

// ScoredRepository 绑定仓库基础信息和对应评分。
type ScoredRepository struct {
	Repository Repository `json:"repository"`
	Score      Score      `json:"score"`
}

// DiscoveryResult 是一次项目发现流程返回给前端和报告服务的完整结果。
type DiscoveryResult struct {
	Intent         SearchIntent       `json:"intent"`
	Queries        []PlannedQuery     `json:"queries"`
	Repositories   []ScoredRepository `json:"repositories"`
	MarkdownReport string             `json:"markdown_report"`
	UsedLiveGitHub bool               `json:"used_live_github"`
	Warnings       []string           `json:"warnings"`
}

package domain

import "time"

// SearchIntent 表示从用户自然语言目标中抽取出的结构化搜索意图。
type SearchIntent struct {
	UserInput   string   `json:"user_input"`
	Languages   []string `json:"languages"`
	Topics      []string `json:"topics"`
	TargetRole  string   `json:"target_role"`
	Goals       []string `json:"goals"`
	Difficulty  string   `json:"difficulty"`
	ProjectSize int      `json:"project_size"`
	MinStars    int      `json:"min_stars"`
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

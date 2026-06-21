package domain

import "time"

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

type PlannedQuery struct {
	Query       string `json:"query"`
	Reason      string `json:"reason"`
	Description string `json:"description"`
}

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

type ScoredRepository struct {
	Repository Repository `json:"repository"`
	Score      Score      `json:"score"`
}

type DiscoveryResult struct {
	Intent         SearchIntent       `json:"intent"`
	Queries        []PlannedQuery     `json:"queries"`
	Repositories   []ScoredRepository `json:"repositories"`
	MarkdownReport string             `json:"markdown_report"`
	UsedLiveGitHub bool               `json:"used_live_github"`
	Warnings       []string           `json:"warnings"`
}

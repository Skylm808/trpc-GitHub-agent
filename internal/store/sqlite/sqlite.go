package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"trpc-GitHub-agent/internal/domain"
)

type SQLiteStore struct {
	db *sql.DB
}

func DefaultPath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "trpc-GitHub-agent", "research.db"), nil
}

func Open(path string) (*SQLiteStore, error) {
	if path == "" {
		var err error
		path, err = DefaultPath()
		if err != nil {
			return nil, err
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	store := &SQLiteStore{db: db}
	if err := store.Migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) Migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, schema)
	return err
}

func (s *SQLiteStore) UpsertRepository(ctx context.Context, repo domain.Repository) (int64, error) {
	topics, _ := json.Marshal(repo.Topics)
	now := time.Now().UTC()
	if repo.FetchedAt.IsZero() {
		repo.FetchedAt = now
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO repositories (
			github_id, full_name, owner, name, description, html_url, clone_url, language,
			topics_json, stars, forks, watchers, open_issues_count, default_branch,
			archived, disabled, pushed_at, created_at, updated_at, fetched_at, raw_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(github_id) DO UPDATE SET
			full_name=excluded.full_name,
			owner=excluded.owner,
			name=excluded.name,
			description=excluded.description,
			html_url=excluded.html_url,
			clone_url=excluded.clone_url,
			language=excluded.language,
			topics_json=excluded.topics_json,
			stars=excluded.stars,
			forks=excluded.forks,
			watchers=excluded.watchers,
			open_issues_count=excluded.open_issues_count,
			default_branch=excluded.default_branch,
			archived=excluded.archived,
			disabled=excluded.disabled,
			pushed_at=excluded.pushed_at,
			created_at=excluded.created_at,
			updated_at=excluded.updated_at,
			fetched_at=excluded.fetched_at,
			raw_json=excluded.raw_json
	`, repo.GitHubID, repo.FullName, repo.Owner, repo.Name, repo.Description, repo.HTMLURL, repo.CloneURL, repo.Language,
		string(topics), repo.Stars, repo.Forks, repo.Watchers, repo.OpenIssuesCount, repo.DefaultBranch,
		repo.Archived, repo.Disabled, repo.PushedAt, repo.CreatedAt, repo.UpdatedAt, repo.FetchedAt, repo.RawJSON)
	if err != nil {
		return 0, err
	}
	return s.repositoryID(ctx, repo.GitHubID)
}

func (s *SQLiteStore) GetRepositoryByFullName(ctx context.Context, fullName string) (domain.Repository, bool, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, github_id, full_name, owner, name, description, html_url, clone_url, language,
			topics_json, stars, forks, watchers, open_issues_count, default_branch,
			archived, disabled, pushed_at, created_at, updated_at, fetched_at, raw_json
		FROM repositories WHERE full_name = ?
	`, fullName)
	repo, err := scanRepository(row)
	if err == sql.ErrNoRows {
		return domain.Repository{}, false, nil
	}
	if err != nil {
		return domain.Repository{}, false, err
	}
	return repo, true, nil
}

func (s *SQLiteStore) SaveScore(ctx context.Context, score domain.Score) error {
	explanation, _ := json.Marshal(score.Explanation)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO project_scores (
			repository_id, activity_score, popularity_score, learning_value_score,
			contribution_friendliness_score, role_relevance_score, total_score,
			influence_level, beginner_friendliness, difficulty, explanation_json, scored_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, score.RepositoryID, score.ActivityScore, score.PopularityScore, score.LearningValueScore,
		score.ContributionFriendlinessScore, score.RoleRelevanceScore, score.TotalScore,
		score.InfluenceLevel, score.BeginnerFriendliness, score.Difficulty, string(explanation), score.ScoredAt)
	return err
}

func (s *SQLiteStore) SaveReport(ctx context.Context, repositoryID int64, reportType, title, markdown string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO reports (repository_id, report_type, title, markdown, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, repositoryID, reportType, title, markdown, time.Now().UTC())
	return err
}

func (s *SQLiteStore) SaveQueryHistory(ctx context.Context, intent domain.SearchIntent, queries []domain.PlannedQuery, repoIDs []int64) error {
	intentJSON, _ := json.Marshal(intent)
	queriesJSON, _ := json.Marshal(queries)
	repoIDsJSON, _ := json.Marshal(repoIDs)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO query_history (user_input, extracted_intent_json, generated_queries_json, result_repository_ids_json, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, intent.UserInput, string(intentJSON), string(queriesJSON), string(repoIDsJSON), time.Now().UTC())
	return err
}

// SaveUserPreference 保存或更新长期偏好，供 Agent 记忆用户技术栈和目标。
func (s *SQLiteStore) SaveUserPreference(ctx context.Context, key string, value any) error {
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO user_preferences (key, value_json, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value_json=excluded.value_json,
			updated_at=excluded.updated_at
	`, key, string(valueJSON), time.Now().UTC())
	return err
}

func (s *SQLiteStore) repositoryID(ctx context.Context, githubID int64) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, `SELECT id FROM repositories WHERE github_id = ?`, githubID).Scan(&id)
	return id, err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRepository(row rowScanner) (domain.Repository, error) {
	var repo domain.Repository
	var topicsJSON string
	err := row.Scan(&repo.ID, &repo.GitHubID, &repo.FullName, &repo.Owner, &repo.Name, &repo.Description,
		&repo.HTMLURL, &repo.CloneURL, &repo.Language, &topicsJSON, &repo.Stars, &repo.Forks,
		&repo.Watchers, &repo.OpenIssuesCount, &repo.DefaultBranch, &repo.Archived, &repo.Disabled,
		&repo.PushedAt, &repo.CreatedAt, &repo.UpdatedAt, &repo.FetchedAt, &repo.RawJSON)
	if err != nil {
		return domain.Repository{}, err
	}
	if strings.TrimSpace(topicsJSON) != "" {
		_ = json.Unmarshal([]byte(topicsJSON), &repo.Topics)
	}
	return repo, nil
}

func AnalyzeTree(paths []string) (hasDocs, hasExamples, hasTests, hasContributing bool, dependencySummary, structureSummary string) {
	var deps []string
	for _, path := range paths {
		lower := strings.ToLower(path)
		switch {
		case lower == "docs" || strings.HasPrefix(lower, "docs/"):
			hasDocs = true
		case lower == "examples" || strings.HasPrefix(lower, "examples/"):
			hasExamples = true
		case strings.Contains(lower, "test") || strings.HasPrefix(lower, "tests/"):
			hasTests = true
		case lower == "contributing.md" || strings.Contains(lower, "/contributing."):
			hasContributing = true
		case lower == "go.mod", lower == "package.json", lower == "pyproject.toml", lower == "requirements.txt", lower == "cargo.toml":
			deps = append(deps, path)
		}
	}
	if len(deps) > 0 {
		dependencySummary = "Dependency files: " + strings.Join(deps, ", ")
	}
	if len(paths) > 0 {
		structureSummary = fmt.Sprintf("Repository tree includes %d tracked paths.", len(paths))
	}
	return
}

const schema = `
CREATE TABLE IF NOT EXISTS repositories (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	github_id INTEGER NOT NULL UNIQUE,
	full_name TEXT NOT NULL UNIQUE,
	owner TEXT NOT NULL,
	name TEXT NOT NULL,
	description TEXT,
	html_url TEXT,
	clone_url TEXT,
	language TEXT,
	topics_json TEXT,
	stars INTEGER NOT NULL DEFAULT 0,
	forks INTEGER NOT NULL DEFAULT 0,
	watchers INTEGER NOT NULL DEFAULT 0,
	open_issues_count INTEGER NOT NULL DEFAULT 0,
	default_branch TEXT,
	archived BOOLEAN NOT NULL DEFAULT FALSE,
	disabled BOOLEAN NOT NULL DEFAULT FALSE,
	pushed_at TIMESTAMP,
	created_at TIMESTAMP,
	updated_at TIMESTAMP,
	fetched_at TIMESTAMP,
	etag TEXT,
	raw_json TEXT
);

CREATE TABLE IF NOT EXISTS repo_snapshots (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	repository_id INTEGER NOT NULL,
	readme_text TEXT,
	tree_json TEXT,
	dependency_files_json TEXT,
	docs_summary TEXT,
	examples_summary TEXT,
	tests_summary TEXT,
	fetched_at TIMESTAMP,
	FOREIGN KEY(repository_id) REFERENCES repositories(id)
);

CREATE TABLE IF NOT EXISTS issues (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	repository_id INTEGER NOT NULL,
	github_issue_id INTEGER,
	number INTEGER,
	title TEXT,
	state TEXT,
	labels_json TEXT,
	author TEXT,
	comments INTEGER,
	created_at TIMESTAMP,
	updated_at TIMESTAMP,
	category TEXT,
	difficulty TEXT,
	beginner_score INTEGER,
	raw_json TEXT,
	FOREIGN KEY(repository_id) REFERENCES repositories(id)
);

CREATE TABLE IF NOT EXISTS pull_requests (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	repository_id INTEGER NOT NULL,
	github_pr_id INTEGER,
	number INTEGER,
	title TEXT,
	state TEXT,
	labels_json TEXT,
	changed_files INTEGER,
	additions INTEGER,
	deletions INTEGER,
	risk_level TEXT,
	summary TEXT,
	raw_json TEXT,
	FOREIGN KEY(repository_id) REFERENCES repositories(id)
);

CREATE TABLE IF NOT EXISTS project_scores (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	repository_id INTEGER NOT NULL,
	activity_score INTEGER,
	popularity_score INTEGER,
	learning_value_score INTEGER,
	contribution_friendliness_score INTEGER,
	role_relevance_score INTEGER,
	total_score INTEGER,
	influence_level TEXT,
	beginner_friendliness TEXT,
	difficulty TEXT,
	explanation_json TEXT,
	scored_at TIMESTAMP,
	FOREIGN KEY(repository_id) REFERENCES repositories(id)
);

CREATE TABLE IF NOT EXISTS reports (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	repository_id INTEGER,
	report_type TEXT,
	title TEXT,
	markdown TEXT,
	created_at TIMESTAMP,
	model_provider TEXT,
	model_name TEXT
);

CREATE TABLE IF NOT EXISTS user_preferences (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	key TEXT NOT NULL UNIQUE,
	value_json TEXT NOT NULL,
	updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS query_history (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_input TEXT NOT NULL,
	extracted_intent_json TEXT,
	generated_queries_json TEXT,
	result_repository_ids_json TEXT,
	created_at TIMESTAMP
);
`

package main

import (
	"context"
	"sort"
	"strings"
	"time"

	"trpc-GitHub-agent/internal/config"
	"trpc-GitHub-agent/internal/domain"
	gh "trpc-GitHub-agent/internal/github"
	"trpc-GitHub-agent/internal/query"
	"trpc-GitHub-agent/internal/report"
	"trpc-GitHub-agent/internal/scoring"
	"trpc-GitHub-agent/internal/store"
)

type App struct {
	ctx       context.Context
	store     *store.SQLiteStore
	github    *gh.Client
	planner   *query.Planner
	scorer    *scoring.Service
	reporter  *report.Service
	storePath string
}

func NewApp() *App {
	return &App{
		github:   gh.NewClient(),
		planner:  query.NewPlanner(),
		scorer:   scoring.NewService(),
		reporter: report.NewService(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	sqliteStore, err := store.Open("")
	if err == nil {
		a.store = sqliteStore
		a.storePath, _ = store.DefaultPath()
	}
}

func (a *App) shutdown(ctx context.Context) {
	if a.store != nil {
		_ = a.store.Close()
	}
}

func (a *App) DiscoverProjects(userInput string, limit int) (domain.DiscoveryResult, error) {
	if strings.TrimSpace(userInput) == "" {
		userInput = "Go backend Agent projects for learning and open source contribution"
	}
	intent, queries := a.planner.Plan(userInput, limit)
	if limit <= 0 {
		limit = intent.ProjectSize
	}
	if limit <= 0 {
		limit = 10
	}

	var warnings []string
	repos, usedLive := a.searchLiveRepositories(queries, limit, &warnings)
	if len(repos) == 0 {
		repos = fallbackRepositories()
		usedLive = false
		warnings = append(warnings, "GitHub live search unavailable or empty; showing deterministic demo fixtures.")
	}

	scored := make([]domain.ScoredRepository, 0, len(repos))
	var repoIDs []int64
	for _, repo := range repos {
		repoID := repo.ID
		if a.store != nil {
			if savedID, err := a.store.UpsertRepository(context.Background(), repo); err == nil {
				repo.ID = savedID
				repoID = savedID
				repoIDs = append(repoIDs, savedID)
			} else {
				warnings = append(warnings, "SQLite cache write failed for "+repo.FullName+": "+err.Error())
			}
		}
		profile := a.profileRepository(repo, usedLive, &warnings)
		profile.Repository.ID = repoID
		score := a.scorer.Score(profile, intent)
		if a.store != nil && repoID > 0 {
			_ = a.store.SaveScore(context.Background(), score)
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
	result.MarkdownReport = a.reporter.Recommendation(result)
	if a.store != nil {
		_ = a.store.SaveQueryHistory(context.Background(), intent, queries, repoIDs)
		if len(scored) > 0 {
			_ = a.store.SaveReport(context.Background(), scored[0].Repository.ID, "recommendation", "GitHub Project Recommendation Report", result.MarkdownReport)
		}
	}
	return result, nil
}

func (a *App) StorePath() string {
	return a.storePath
}

func (a *App) SettingsStatus() config.SettingsStatus {
	return config.LoadSettingsStatus()
}

func (a *App) searchLiveRepositories(queries []domain.PlannedQuery, limit int, warnings *[]string) ([]domain.Repository, bool) {
	if a.github == nil {
		return nil, false
	}
	seen := map[string]bool{}
	var repos []domain.Repository
	perQuery := limit
	if len(queries) > 1 {
		perQuery = max(3, limit/len(queries)+1)
	}
	for _, planned := range queries {
		found, err := a.github.SearchRepositories(context.Background(), planned.Query, perQuery)
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

func (a *App) profileRepository(repo domain.Repository, live bool, warnings *[]string) domain.RepositoryProfile {
	profile := domain.RepositoryProfile{
		Repository:       repo,
		HasReadme:        true,
		StructureSummary: "Repository metadata available from GitHub search.",
	}
	if !live || a.github == nil {
		applyFixtureSignals(&profile)
		return profile
	}

	readme, err := a.github.GetReadme(context.Background(), repo.FullName)
	if err == nil && strings.TrimSpace(readme) != "" {
		profile.HasReadme = true
		profile.ReadmeSummary = summarizeReadme(readme)
	} else if err != nil {
		*warnings = append(*warnings, "README fetch failed for "+repo.FullName+": "+err.Error())
	}

	tree, err := a.github.GetTree(context.Background(), repo.FullName, repo.DefaultBranch)
	if err == nil {
		paths := make([]string, 0, len(tree))
		for _, item := range tree {
			paths = append(paths, item.Path)
		}
		hasDocs, hasExamples, hasTests, hasContributing, dependencySummary, structureSummary := store.AnalyzeTree(paths)
		profile.HasDocs = hasDocs
		profile.HasExamples = hasExamples
		profile.HasTests = hasTests
		profile.HasContributing = hasContributing
		profile.DependencySummary = dependencySummary
		profile.StructureSummary = structureSummary
	} else {
		*warnings = append(*warnings, "Tree fetch failed for "+repo.FullName+": "+err.Error())
	}

	if count, err := a.github.CountIssuesByLabel(context.Background(), repo.FullName, "good first issue"); err == nil {
		profile.GoodFirstIssueCount = count
	}
	if count, err := a.github.CountIssuesByLabel(context.Background(), repo.FullName, "help wanted"); err == nil {
		profile.HelpWantedCount = count
	}
	return profile
}

func summarizeReadme(readme string) string {
	readme = strings.TrimSpace(readme)
	if len(readme) <= 280 {
		return readme
	}
	return readme[:280] + "..."
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

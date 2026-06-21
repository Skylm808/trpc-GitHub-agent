package scoring

import (
	"testing"
	"time"

	"trpc-GitHub-agent/internal/domain"
)

func TestActiveRepositoryScoresHigherThanStaleRepository(t *testing.T) {
	now := time.Date(2026, 6, 22, 0, 0, 0, 0, time.UTC)
	service := NewServiceWithClock(func() time.Time { return now })
	intent := domain.SearchIntent{Languages: []string{"Go"}, Topics: []string{"agent"}, Difficulty: "intermediate"}

	active := profileWithRepo("owner/active", 500, now.AddDate(0, 0, -5))
	stale := profileWithRepo("owner/stale", 500, now.AddDate(-2, 0, 0))

	activeScore := service.Score(active, intent)
	staleScore := service.Score(stale, intent)

	if activeScore.ActivityScore <= staleScore.ActivityScore {
		t.Fatalf("expected active activity score > stale, got %d <= %d", activeScore.ActivityScore, staleScore.ActivityScore)
	}
	if activeScore.TotalScore <= staleScore.TotalScore {
		t.Fatalf("expected active total score > stale, got %d <= %d", activeScore.TotalScore, staleScore.TotalScore)
	}
}

func TestLearningAndContributionSignalsIncreaseScore(t *testing.T) {
	now := time.Date(2026, 6, 22, 0, 0, 0, 0, time.UTC)
	service := NewServiceWithClock(func() time.Time { return now })
	intent := domain.SearchIntent{Languages: []string{"Go"}, Topics: []string{"agent"}, Difficulty: "beginner"}

	bare := profileWithRepo("owner/bare", 200, now)
	bare.HasDocs = false
	bare.HasExamples = false
	bare.HasTests = false
	bare.HasContributing = false
	bare.GoodFirstIssueCount = 0

	friendly := profileWithRepo("owner/friendly", 200, now)
	friendly.HasDocs = true
	friendly.HasExamples = true
	friendly.HasTests = true
	friendly.HasContributing = true
	friendly.GoodFirstIssueCount = 4

	bareScore := service.Score(bare, intent)
	friendlyScore := service.Score(friendly, intent)

	if friendlyScore.LearningValueScore <= bareScore.LearningValueScore {
		t.Fatalf("expected learning value to increase, got %d <= %d", friendlyScore.LearningValueScore, bareScore.LearningValueScore)
	}
	if friendlyScore.ContributionFriendlinessScore <= bareScore.ContributionFriendlinessScore {
		t.Fatalf("expected contribution score to increase, got %d <= %d", friendlyScore.ContributionFriendlinessScore, bareScore.ContributionFriendlinessScore)
	}
	if friendlyScore.BeginnerFriendliness != "beginner-friendly" {
		t.Fatalf("expected beginner-friendly, got %q", friendlyScore.BeginnerFriendliness)
	}
}

func profileWithRepo(fullName string, stars int, pushedAt time.Time) domain.RepositoryProfile {
	return domain.RepositoryProfile{
		Repository: domain.Repository{
			FullName:        fullName,
			Description:     "Go agent framework",
			Language:        "Go",
			Topics:          []string{"agent"},
			Stars:           stars,
			Forks:           20,
			OpenIssuesCount: 5,
			PushedAt:        pushedAt,
			UpdatedAt:       pushedAt,
		},
		HasReadme:         true,
		HasDocs:           true,
		HasExamples:       true,
		HasTests:          true,
		HasContributing:   true,
		DependencySummary: "Dependency files: go.mod",
		StructureSummary:  "Repository tree includes 20 tracked paths.",
	}
}

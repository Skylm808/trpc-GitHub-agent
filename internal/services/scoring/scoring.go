package scoring

import (
	"fmt"
	"math"
	"strings"
	"time"

	"trpc-GitHub-agent/internal/domain"
)

type Service struct {
	now func() time.Time
}

func NewService() *Service {
	return &Service{now: time.Now}
}

func NewServiceWithClock(now func() time.Time) *Service {
	return &Service{now: now}
}

func (s *Service) Score(profile domain.RepositoryProfile, intent domain.SearchIntent) domain.Score {
	repo := profile.Repository
	activity := s.activityScore(repo)
	popularity := popularityScore(repo)
	learning := learningValueScore(profile)
	contribution := contributionScore(profile)
	relevance := roleRelevanceScore(repo, intent)
	total := activity + popularity + learning + contribution + relevance
	beginner := beginnerFriendliness(profile)
	difficulty := difficulty(profile, repo, intent)
	influence := influenceLevel(repo.Stars)

	return domain.Score{
		RepositoryID:                  repo.ID,
		ActivityScore:                 activity,
		PopularityScore:               popularity,
		LearningValueScore:            learning,
		ContributionFriendlinessScore: contribution,
		RoleRelevanceScore:            relevance,
		TotalScore:                    total,
		InfluenceLevel:                influence,
		BeginnerFriendliness:          beginner,
		Difficulty:                    difficulty,
		RecommendationReason:          recommendation(repo, total, influence, beginner, difficulty),
		Explanation: map[string]string{
			"activity":                  fmt.Sprintf("Activity score %d/20 based on pushed_at and repository availability.", activity),
			"popularity":                fmt.Sprintf("Popularity score %d/15 based on stars and forks with log normalization.", popularity),
			"learning_value":            fmt.Sprintf("Learning value score %d/25 based on README, docs, examples, tests, dependencies, and structure.", learning),
			"contribution_friendliness": fmt.Sprintf("Contribution score %d/20 based on good-first/help-wanted issues, contribution guide, tests, and recent issue activity.", contribution),
			"role_relevance":            fmt.Sprintf("Role relevance score %d/20 based on language and topic match.", relevance),
		},
		ScoredAt: s.now(),
	}
}

func (s *Service) activityScore(repo domain.Repository) int {
	score := 0
	age := s.now().Sub(repo.PushedAt)
	switch {
	case repo.PushedAt.IsZero():
		score += 1
	case age <= 30*24*time.Hour:
		score += 10
	case age <= 90*24*time.Hour:
		score += 7
	case age <= 180*24*time.Hour:
		score += 4
	default:
		score += 1
	}
	if repo.OpenIssuesCount > 0 {
		score += 5
	}
	if !repo.Archived && !repo.Disabled {
		score += 2
	}
	if !repo.UpdatedAt.IsZero() && s.now().Sub(repo.UpdatedAt) <= 90*24*time.Hour {
		score += 3
	}
	return min(score, 20)
}

func popularityScore(repo domain.Repository) int {
	score := 0
	if repo.Stars > 0 {
		score += int(math.Round(math.Log10(float64(repo.Stars)+1) / math.Log10(30001) * 12))
	}
	if repo.Forks > 0 {
		score += min(3, int(math.Round(math.Log10(float64(repo.Forks)+1))))
	}
	return min(score, 15)
}

func learningValueScore(profile domain.RepositoryProfile) int {
	score := 0
	if profile.HasReadme {
		score += 6
	}
	if profile.HasExamples {
		score += 5
	}
	if profile.HasDocs {
		score += 4
	}
	if profile.HasTests {
		score += 4
	}
	if profile.DependencySummary != "" {
		score += 3
	}
	if profile.StructureSummary != "" {
		score += 3
	}
	return min(score, 25)
}

func contributionScore(profile domain.RepositoryProfile) int {
	score := 0
	if profile.GoodFirstIssueCount > 0 {
		score += min(6, 2+profile.GoodFirstIssueCount)
	}
	if profile.HelpWantedCount > 0 {
		score += min(4, 1+profile.HelpWantedCount)
	}
	if profile.HasContributing {
		score += 3
	}
	if profile.HasTests {
		score += 3
	}
	if profile.Repository.OpenIssuesCount > 0 && !profile.Repository.UpdatedAt.IsZero() {
		score += 4
	}
	return min(score, 20)
}

func roleRelevanceScore(repo domain.Repository, intent domain.SearchIntent) int {
	score := 0
	for _, language := range intent.Languages {
		if strings.EqualFold(repo.Language, language) {
			score += 5
			break
		}
	}
	text := strings.ToLower(repo.FullName + " " + repo.Description + " " + strings.Join(repo.Topics, " "))
	topicMatches := 0
	for _, topic := range intent.Topics {
		if strings.Contains(text, strings.ToLower(topic)) {
			topicMatches++
		}
	}
	score += min(7, topicMatches*3)
	if intent.TargetRole != "" && strings.Contains(text, strings.ToLower(intent.TargetRole)) {
		score += 3
	}
	if intent.Difficulty != "" {
		score += 2
	}
	return min(score, 20)
}

func influenceLevel(stars int) string {
	switch {
	case stars >= 30000:
		return "S"
	case stars >= 10000:
		return "A"
	case stars >= 1000:
		return "B"
	case stars >= 100:
		return "C"
	default:
		return "D"
	}
}

func beginnerFriendliness(profile domain.RepositoryProfile) string {
	signals := 0
	for _, enabled := range []bool{profile.HasReadme, profile.HasDocs, profile.HasExamples, profile.HasTests, profile.HasContributing, profile.GoodFirstIssueCount > 0} {
		if enabled {
			signals++
		}
	}
	switch {
	case profile.Repository.Archived || profile.Repository.Disabled:
		return "not-recommended"
	case signals >= 5:
		return "beginner-friendly"
	case signals >= 3:
		return "intermediate-friendly"
	default:
		return "advanced-or-unclear"
	}
}

func difficulty(profile domain.RepositoryProfile, repo domain.Repository, intent domain.SearchIntent) string {
	text := strings.ToLower(repo.FullName + " " + repo.Description + " " + strings.Join(repo.Topics, " "))
	if strings.Contains(text, "graph") || strings.Contains(text, "memory") || strings.Contains(text, "evaluation") || strings.Contains(text, "runtime") {
		return "advanced"
	}
	if beginnerFriendliness(profile) == "beginner-friendly" && intent.Difficulty == "beginner" {
		return "beginner"
	}
	if profile.GoodFirstIssueCount > 0 && profile.HasExamples && profile.HasTests {
		return "intermediate"
	}
	return "intermediate"
}

func recommendation(repo domain.Repository, total int, influence, beginner, difficulty string) string {
	switch {
	case influence == "S" || influence == "A":
		return fmt.Sprintf("%s 影响力较强，总分 %d/100，更适合源码阅读、简历分析和重点模块研究；贡献建议从范围清晰的 Issue 开始。", repo.FullName, total)
	case beginner == "beginner-friendly":
		return fmt.Sprintf("%s 总分 %d/100，并具备新手友好信号，适合作为首次开源贡献路径。", repo.FullName, total)
	default:
		return fmt.Sprintf("%s 总分 %d/100，难度为 %s、影响力为 %s；如果与当前技术栈匹配，适合做定向学习和模块拆解。", repo.FullName, total, difficulty, influence)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

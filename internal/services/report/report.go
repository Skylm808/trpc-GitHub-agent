package report

import (
	"fmt"
	"strings"

	"trpc-GitHub-agent/internal/domain"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Recommendation(result domain.DiscoveryResult) string {
	var b strings.Builder
	b.WriteString("# GitHub Project Recommendation Report\n\n")
	b.WriteString("## User Intent\n\n")
	b.WriteString(fmt.Sprintf("- Input: %s\n", result.Intent.UserInput))
	b.WriteString(fmt.Sprintf("- Languages: %s\n", strings.Join(result.Intent.Languages, ", ")))
	b.WriteString(fmt.Sprintf("- Topics: %s\n", strings.Join(result.Intent.Topics, ", ")))
	b.WriteString(fmt.Sprintf("- Target role: %s\n", result.Intent.TargetRole))
	b.WriteString(fmt.Sprintf("- Difficulty: %s\n\n", result.Intent.Difficulty))

	b.WriteString("## Generated GitHub Queries\n\n")
	for _, query := range result.Queries {
		b.WriteString(fmt.Sprintf("- `%s`\n  - %s\n", query.Query, query.Reason))
	}
	b.WriteString("\n## Ranked Projects\n\n")
	b.WriteString("| Rank | Repository | Score | Influence | Friendly | Difficulty | Reason |\n")
	b.WriteString("| --- | --- | ---: | --- | --- | --- | --- |\n")
	for i, scored := range result.Repositories {
		repo := scored.Repository
		score := scored.Score
		b.WriteString(fmt.Sprintf("| %d | [%s](%s) | %d | %s | %s | %s | %s |\n",
			i+1, repo.FullName, repo.HTMLURL, score.TotalScore, score.InfluenceLevel,
			score.BeginnerFriendliness, score.Difficulty, escapeTable(score.RecommendationReason)))
	}
	b.WriteString("\n## Notes\n\n")
	if result.UsedLiveGitHub {
		b.WriteString("- Data source: live GitHub REST API.\n")
	} else {
		b.WriteString("- Data source: local fixtures or cached fallback.\n")
	}
	if len(result.Warnings) > 0 {
		for _, warning := range result.Warnings {
			b.WriteString(fmt.Sprintf("- Warning: %s\n", warning))
		}
	}
	return b.String()
}

func escapeTable(value string) string {
	return strings.ReplaceAll(value, "|", "\\|")
}

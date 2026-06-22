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
	b.WriteString("# GitHub 开源项目研究报告\n\n")
	b.WriteString("## 用户目标\n\n")
	b.WriteString(fmt.Sprintf("- 输入：%s\n", result.Intent.UserInput))
	b.WriteString(fmt.Sprintf("- 编程语言：%s\n", strings.Join(result.Intent.Languages, ", ")))
	b.WriteString(fmt.Sprintf("- 技术方向：%s\n", strings.Join(result.Intent.Topics, ", ")))
	b.WriteString(fmt.Sprintf("- 目标角色：%s\n", result.Intent.TargetRole))
	b.WriteString(fmt.Sprintf("- 难度：%s\n\n", result.Intent.Difficulty))

	b.WriteString("## 生成的 GitHub 查询\n\n")
	for _, query := range result.Queries {
		b.WriteString(fmt.Sprintf("- `%s`\n  - %s\n", query.Query, query.Reason))
	}
	b.WriteString("\n## 项目分级结果\n\n")
	b.WriteString("| 排名 | 仓库 | 总分 | 影响力 | 新手友好 | 难度 | 推荐理由 |\n")
	b.WriteString("| --- | --- | ---: | --- | --- | --- | --- |\n")
	for i, scored := range result.Repositories {
		repo := scored.Repository
		score := scored.Score
		b.WriteString(fmt.Sprintf("| %d | [%s](%s) | %d | %s | %s | %s | %s |\n",
			i+1, repo.FullName, repo.HTMLURL, score.TotalScore, score.InfluenceLevel,
			score.BeginnerFriendliness, score.Difficulty, escapeTable(score.RecommendationReason)))
	}
	b.WriteString("\n## 说明\n\n")
	if result.UsedLiveGitHub {
		b.WriteString("- 数据来源：GitHub REST API 实时数据。\n")
	} else {
		b.WriteString("- 数据来源：本地 fixture 或缓存兜底数据。\n")
	}
	if len(result.Warnings) > 0 {
		for _, warning := range result.Warnings {
			b.WriteString(fmt.Sprintf("- 注意：%s\n", warning))
		}
	}
	return b.String()
}

func escapeTable(value string) string {
	return strings.ReplaceAll(value, "|", "\\|")
}

package queryplanner

import (
	"unicode"

	"strings"

	"trpc-GitHub-agent/internal/domain"
)

type Planner struct{}

func NewPlanner() *Planner {
	return &Planner{}
}

func (p *Planner) Plan(input string, limit int, filters domain.SearchIntent) (domain.SearchIntent, []domain.PlannedQuery) {
	normalized := strings.ToLower(input)
	intent := domain.SearchIntent{
		UserInput:     input,
		InputLanguage: filters.InputLanguage,
		Languages:     detectLanguages(normalized),
		Topics:        detectTopics(normalized),
		TargetRole:    detectRole(normalized),
		Goals:         detectGoals(normalized),
		Difficulty:    detectDifficulty(normalized),
		Direction:     filters.Direction,
		PushedAfter:   filters.PushedAfter,
		ProjectSize:   limit,
		MinStars:      filters.MinStars,
		MaxStars:      filters.MaxStars,
	}
	if intent.ProjectSize <= 0 {
		intent.ProjectSize = 10
	}
	intent = mergeFilters(intent, filters)
	if intent.MinStars <= 0 {
		intent.MinStars = minStars(intent)
	}
	if intent.MaxStars <= 0 {
		intent.MaxStars = maxStars(intent)
	}

	queries := buildQueries(intent)
	return intent, queries
}

func detectLanguages(input string) []string {
	var langs []string
	candidates := map[string]string{
		"go":         "Go",
		"golang":     "Go",
		"后端":         "Go",
		"python":     "Python",
		"爬虫":         "Python",
		"typescript": "TypeScript",
		"ts":         "TypeScript",
		"javascript": "JavaScript",
		"js":         "JavaScript",
		"java":       "Java",
		"rust":       "Rust",
	}
	for key, lang := range candidates {
		if containsTerm(input, key) && !contains(langs, lang) {
			langs = append(langs, lang)
		}
	}
	if len(langs) == 0 {
		langs = append(langs, "Go")
	}
	return langs
}

func detectTopics(input string) []string {
	var topics []string
	candidates := map[string][]string{
		"agent":     {"agent", "ai-agent", "llm-agent"},
		"智能体":       {"agent", "ai-agent", "llm-agent"},
		"大模型":       {"llm-agent"},
		"rag":       {"rag", "retrieval"},
		"检索":        {"rag", "retrieval"},
		"mcp":       {"mcp", "model-context-protocol"},
		"协议":        {"mcp"},
		"后端":        {"backend"},
		"backend":   {"backend"},
		"框架":        {"framework"},
		"framework": {"framework"},
	}
	for key, values := range candidates {
		if strings.Contains(input, key) {
			for _, value := range values {
				if !contains(topics, value) {
					topics = append(topics, value)
				}
			}
		}
	}
	if len(topics) == 0 {
		topics = append(topics, "agent")
	}
	return topics
}

func detectRole(input string) string {
	switch {
	case strings.Contains(input, "后端"), strings.Contains(input, "backend"):
		return "backend"
	case strings.Contains(input, "算法"), strings.Contains(input, "ml"), strings.Contains(input, "ai"):
		return "ai"
	case strings.Contains(input, "前端"), strings.Contains(input, "frontend"):
		return "frontend"
	default:
		return "software-engineer"
	}
}

func detectGoals(input string) []string {
	var goals []string
	candidates := map[string]string{
		"秋招":           "recruiting",
		"面试":           "recruiting",
		"简历":           "recruiting",
		"recruit":      "recruiting",
		"学习":           "learning",
		"learn":        "learning",
		"开源":           "contribution",
		"贡献":           "contribution",
		"contribution": "contribution",
	}
	for key, goal := range candidates {
		if strings.Contains(input, key) && !contains(goals, goal) {
			goals = append(goals, goal)
		}
	}
	if len(goals) == 0 {
		goals = []string{"learning", "contribution"}
	}
	return goals
}

func detectDifficulty(input string) string {
	switch {
	case strings.Contains(input, "入门"), strings.Contains(input, "新手"), strings.Contains(input, "beginner"):
		return "beginner"
	case strings.Contains(input, "进阶"), strings.Contains(input, "高阶"), strings.Contains(input, "advanced"):
		return "advanced"
	default:
		return "intermediate"
	}
}

func minStars(intent domain.SearchIntent) int {
	for _, topic := range intent.Topics {
		if topic == "mcp" || topic == "model-context-protocol" {
			return 50
		}
	}
	if contains(intent.Goals, "contribution") {
		return 100
	}
	return 300
}

func buildQueries(intent domain.SearchIntent) []domain.PlannedQuery {
	var queries []domain.PlannedQuery
	language := intent.Languages[0]
	freshness := pushedAfterFilter(intent.PushedAfter)
	baseParts := []string{"language:" + language, starRangeQuery(intent.MinStars, intent.MaxStars), freshness, "archived:false"}
	if intent.Direction != "" {
		baseParts = append(baseParts, "topic:"+intent.Direction)
	}
	base := strings.Join(compact(baseParts), " ")

	topicGroups := [][]string{
		filterTopics(intent.Topics, "agent", "ai-agent", "llm-agent"),
		filterTopics(intent.Topics, "rag", "retrieval"),
		filterTopics(intent.Topics, "mcp", "model-context-protocol"),
	}
	for _, group := range topicGroups {
		if len(group) == 0 {
			continue
		}
		queries = append(queries, domain.PlannedQuery{
			Query:       strings.Join(group, " OR ") + " " + base,
			Reason:      "匹配 " + strings.Join(group, "/") + " 方向，并限制为活跃的 " + language + " 仓库。",
			Description: "根据用户语言、技术方向和活跃项目偏好生成。",
		})
	}
	if len(queries) == 0 {
		queries = append(queries, domain.PlannedQuery{
			Query:       strings.Join(intent.Topics, " OR ") + " " + base,
			Reason:      "针对识别到的项目主题生成宽泛兜底查询。",
			Description: "未识别到专门主题组时生成。",
		})
	}
	if contains(intent.Goals, "contribution") {
		queries = append(queries, domain.PlannedQuery{
			Query:       strings.Join(intent.Topics, " OR ") + " language:" + language + " " + starRangeQuery(max(50, intent.MinStars), intent.MaxStars) + " good-first-issues:>0 archived:false",
			Reason:      "优先寻找带 good-first issues 的仓库，方便作为开源贡献入口。",
			Description: "根据用户的开源贡献目标生成。",
		})
	}
	return dedupeQueries(queries)
}

func mergeFilters(intent domain.SearchIntent, filters domain.SearchIntent) domain.SearchIntent {
	if len(filters.Languages) > 0 {
		intent.Languages = filters.Languages
	}
	if len(filters.Topics) > 0 {
		intent.Topics = filters.Topics
	}
	if filters.TargetRole != "" {
		intent.TargetRole = filters.TargetRole
	}
	if filters.Difficulty != "" {
		intent.Difficulty = filters.Difficulty
	}
	if filters.Direction != "" {
		intent.Direction = filters.Direction
	}
	if filters.PushedAfter != "" {
		intent.PushedAfter = filters.PushedAfter
	}
	if filters.MinStars > 0 {
		intent.MinStars = filters.MinStars
	}
	if filters.MaxStars > 0 {
		intent.MaxStars = filters.MaxStars
	}
	if filters.InputLanguage != "" {
		intent.InputLanguage = filters.InputLanguage
	}
	return intent
}

func pushedAfterFilter(value string) string {
	if strings.TrimSpace(value) == "" {
		return "pushed:>2025-01-01"
	}
	return "pushed:>" + strings.TrimSpace(value)
}

func starRangeQuery(minStars, maxStars int) string {
	if minStars <= 0 && maxStars <= 0 {
		return "stars:>100"
	}
	if minStars > 0 && maxStars > 0 {
		return "stars:" + itoa(minStars) + ".." + itoa(maxStars)
	}
	if minStars > 0 {
		return "stars:>=" + itoa(minStars)
	}
	return "stars:<=" + itoa(maxStars)
}

func maxStars(intent domain.SearchIntent) int {
	if contains(intent.Goals, "contribution") {
		return 50000
	}
	return 100000
}

func compact(values []string) []string {
	var out []string
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, value)
		}
	}
	return out
}

func filterTopics(topics []string, values ...string) []string {
	var out []string
	for _, value := range values {
		if contains(topics, value) {
			out = append(out, value)
		}
	}
	return out
}

func dedupeQueries(queries []domain.PlannedQuery) []domain.PlannedQuery {
	seen := map[string]bool{}
	var out []domain.PlannedQuery
	for _, query := range queries {
		if seen[query.Query] {
			continue
		}
		seen[query.Query] = true
		out = append(out, query)
	}
	return out
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}

func containsTerm(input, term string) bool {
	start := 0
	for {
		index := strings.Index(input[start:], term)
		if index < 0 {
			return false
		}
		index += start
		beforeOK := index == 0 || !isASCIILetterOrDigit(rune(input[index-1]))
		afterIndex := index + len(term)
		afterOK := afterIndex >= len(input) || !isASCIILetterOrDigit(rune(input[afterIndex]))
		if beforeOK && afterOK {
			return true
		}
		start = index + len(term)
	}
}

func isASCIILetterOrDigit(r rune) bool {
	return r <= unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r))
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	var digits []byte
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}

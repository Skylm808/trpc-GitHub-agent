package query

import (
	"strings"

	"trpc-GitHub-agent/internal/domain"
)

type Planner struct{}

func NewPlanner() *Planner {
	return &Planner{}
}

func (p *Planner) Plan(input string, limit int) (domain.SearchIntent, []domain.PlannedQuery) {
	normalized := strings.ToLower(input)
	intent := domain.SearchIntent{
		UserInput:   input,
		Languages:   detectLanguages(normalized),
		Topics:      detectTopics(normalized),
		TargetRole:  detectRole(normalized),
		Goals:       detectGoals(normalized),
		Difficulty:  detectDifficulty(normalized),
		ProjectSize: limit,
	}
	if intent.ProjectSize <= 0 {
		intent.ProjectSize = 10
	}
	intent.MinStars = minStars(intent)

	queries := buildQueries(intent)
	return intent, queries
}

func detectLanguages(input string) []string {
	var langs []string
	candidates := map[string]string{
		"go":         "Go",
		"golang":     "Go",
		"python":     "Python",
		"typescript": "TypeScript",
		"ts":         "TypeScript",
		"javascript": "JavaScript",
		"js":         "JavaScript",
		"java":       "Java",
		"rust":       "Rust",
	}
	for key, lang := range candidates {
		if strings.Contains(input, key) && !contains(langs, lang) {
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
		"rag":       {"rag", "retrieval"},
		"mcp":       {"mcp", "model-context-protocol"},
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
	case strings.Contains(input, "进阶"), strings.Contains(input, "advanced"):
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
	freshness := "pushed:>2025-01-01"
	base := "language:" + language + " stars:>" + itoa(intent.MinStars) + " " + freshness + " archived:false"

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
			Reason:      "Matches " + strings.Join(group, "/") + " with active " + language + " repositories.",
			Description: "Generated from the user's language, direction, and active-project preference.",
		})
	}
	if len(queries) == 0 {
		queries = append(queries, domain.PlannedQuery{
			Query:       strings.Join(intent.Topics, " OR ") + " " + base,
			Reason:      "Broad fallback query for the detected project topics.",
			Description: "Generated because no specialized topic group was detected.",
		})
	}
	if contains(intent.Goals, "contribution") {
		queries = append(queries, domain.PlannedQuery{
			Query:       strings.Join(intent.Topics, " OR ") + " language:" + language + " stars:>50 good-first-issues:>0 archived:false",
			Reason:      "Finds repositories with good-first issues for contribution entry points.",
			Description: "Generated from the user's open source contribution goal.",
		})
	}
	return dedupeQueries(queries)
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

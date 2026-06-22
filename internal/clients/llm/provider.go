package llm

type Provider struct {
	Name        string `json:"name"`
	Model       string `json:"model"`
	BaseURL     string `json:"base_url"`
	Enabled     bool   `json:"enabled"`
	Configured  bool   `json:"configured"`
	SummaryOnly bool   `json:"summary_only"`
}

type ConfiguredProvider struct {
	Name    string
	Model   string
	BaseURL string
	Enabled bool
}

func Providers(activeProvider, activeModel string, configured []ConfiguredProvider, tokenConfigured map[string]bool) []Provider {
	providers := make([]Provider, 0, len(configured))
	for _, provider := range configured {
		model := provider.Model
		if activeProvider == provider.Name && activeModel != "" {
			model = activeModel
		}
		providers = append(providers, Provider{
			Name:        provider.Name,
			Model:       model,
			BaseURL:     provider.BaseURL,
			Enabled:     provider.Enabled,
			Configured:  tokenConfigured[provider.Name],
			SummaryOnly: true,
		})
	}
	if len(providers) == 0 {
		return []Provider{
			{Name: "openai", Model: modelFor(activeProvider, activeModel, "openai", "gpt-4.1-mini"), BaseURL: "https://api.openai.com/v1", Enabled: true, Configured: tokenConfigured["openai"], SummaryOnly: true},
			{Name: "anthropic", Model: modelFor(activeProvider, activeModel, "anthropic", "claude-sonnet"), BaseURL: "https://api.anthropic.com", Enabled: true, Configured: tokenConfigured["anthropic"], SummaryOnly: true},
			{Name: "deepseek", Model: modelFor(activeProvider, activeModel, "deepseek", "deepseek-chat"), BaseURL: "https://api.deepseek.com/v1", Enabled: true, Configured: tokenConfigured["deepseek"], SummaryOnly: true},
			{Name: "custom", Model: modelFor(activeProvider, activeModel, "custom", "gpt-4.1-mini"), Enabled: false, Configured: tokenConfigured["custom"], SummaryOnly: true},
		}
	}
	return providers
}

func modelFor(activeProvider, activeModel, provider, fallback string) string {
	if activeProvider == provider && activeModel != "" {
		return activeModel
	}
	return fallback
}

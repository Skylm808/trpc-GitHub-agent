package config

import "os"

type SettingsStatus struct {
	GitHubTokenConfigured  bool   `json:"github_token_configured"`
	ModelProvider          string `json:"model_provider"`
	ModelName              string `json:"model_name"`
	OpenAIConfigured       bool   `json:"openai_configured"`
	AnthropicConfigured    bool   `json:"anthropic_configured"`
	DeepSeekConfigured     bool   `json:"deepseek_configured"`
	ActiveProviderReady    bool   `json:"active_provider_ready"`
	DeterministicModeReady bool   `json:"deterministic_mode_ready"`
}

func LoadSettingsStatus() SettingsStatus {
	provider := getenv("MODEL_PROVIDER", "openai")
	status := SettingsStatus{
		GitHubTokenConfigured:  os.Getenv("GITHUB_TOKEN") != "",
		ModelProvider:          provider,
		ModelName:              getenv("MODEL_NAME", defaultModel(provider)),
		OpenAIConfigured:       os.Getenv("OPENAI_API_KEY") != "",
		AnthropicConfigured:    os.Getenv("ANTHROPIC_API_KEY") != "",
		DeepSeekConfigured:     os.Getenv("DEEPSEEK_API_KEY") != "",
		DeterministicModeReady: true,
	}
	status.ActiveProviderReady = activeProviderReady(status)
	return status
}

func activeProviderReady(status SettingsStatus) bool {
	switch status.ModelProvider {
	case "openai":
		return status.OpenAIConfigured
	case "anthropic", "claude":
		return status.AnthropicConfigured
	case "deepseek":
		return status.DeepSeekConfigured
	default:
		return false
	}
}

func defaultModel(provider string) string {
	switch provider {
	case "anthropic", "claude":
		return "claude-sonnet"
	case "deepseek":
		return "deepseek-chat"
	default:
		return "gpt-4.1-mini"
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"trpc-GitHub-agent/internal/clients/llm"
)

type SettingsStatus struct {
	GitHubTokenConfigured  bool           `json:"github_token_configured"`
	GitHubBaseURL          string         `json:"github_base_url"`
	ModelProvider          string         `json:"model_provider"`
	ModelName              string         `json:"model_name"`
	OpenAIConfigured       bool           `json:"openai_configured"`
	AnthropicConfigured    bool           `json:"anthropic_configured"`
	DeepSeekConfigured     bool           `json:"deepseek_configured"`
	ActiveProviderReady    bool           `json:"active_provider_ready"`
	DeterministicModeReady bool           `json:"deterministic_mode_ready"`
	Providers              []llm.Provider `json:"providers"`
}

type ConnectionCheck struct {
	Target  string `json:"target"`
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

type ProviderConnectionRequest struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	BaseURL  string `json:"base_url"`
	Token    string `json:"token"`
	Enabled  bool   `json:"enabled"`
}

type GitHubConnectionRequest struct {
	BaseURL string `json:"base_url"`
	Token   string `json:"token"`
}

type RuntimeProvider struct {
	Provider ProviderConfig
	Token    string
}

func LoadSettingsStatus() SettingsStatus {
	cfg, _, _ := loadAppConfig()
	secrets, _, _ := loadSecretConfig()
	return loadSettingsStatus(cfg, secrets)
}

func loadSettingsStatus(cfg AppConfig, secrets SecretConfig) SettingsStatus {
	provider := envOrValue("MODEL_PROVIDER", cfg.ModelProvider)
	if provider == "" {
		provider = "openai"
	}
	provider = normalizeProviderName(provider)
	modelName := envOrValue("MODEL_NAME", cfg.ModelName)
	if modelName == "" {
		modelName = defaultModel(provider)
	}
	githubBaseURL := envOrValue("GITHUB_BASE_URL", cfg.GitHubBaseURL)
	status := SettingsStatus{
		GitHubTokenConfigured:  envOrFile("GITHUB_TOKEN", secrets.GitHubToken),
		GitHubBaseURL:          githubBaseURL,
		ModelProvider:          provider,
		ModelName:              modelName,
		OpenAIConfigured:       providerTokenConfigured("openai", secrets),
		AnthropicConfigured:    providerTokenConfigured("anthropic", secrets),
		DeepSeekConfigured:     providerTokenConfigured("deepseek", secrets),
		DeterministicModeReady: true,
	}
	status.Providers = llm.Providers(status.ModelProvider, status.ModelName, toLLMProviders(cfg.Providers), providerTokenMap(secrets))
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
		for _, provider := range status.Providers {
			if provider.Name == status.ModelProvider {
				return provider.Configured
			}
		}
		return false
	}
}

func ResolvedGitHubBaseURL() string {
	cfg, _, _ := loadAppConfig()
	return envOrValue("GITHUB_BASE_URL", cfg.GitHubBaseURL)
}

func ResolvedGitHubToken() string {
	_, _, _ = loadAppConfig()
	secrets, _, _ := loadSecretConfig()
	if value := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); value != "" {
		return value
	}
	return strings.TrimSpace(secrets.GitHubToken)
}

func CheckGitHubConnection() ConnectionCheck {
	baseURL := ResolvedGitHubBaseURL()
	if strings.TrimSpace(baseURL) == "" {
		return ConnectionCheck{Target: "github", OK: false, Message: "GitHub API base URL 未配置"}
	}
	if strings.TrimSpace(ResolvedGitHubToken()) == "" {
		return ConnectionCheck{Target: "github", OK: false, Message: "GitHub token 未配置，公共 API 仍可用但会很快限流"}
	}
	return ConnectionCheck{Target: "github", OK: true, Message: "GitHub base URL 与 token 已配置"}
}

func CheckGitHubConnectionRequest(request GitHubConnectionRequest) ConnectionCheck {
	baseURL := strings.TrimRight(strings.TrimSpace(request.BaseURL), "/")
	if baseURL == "" {
		return ConnectionCheck{Target: "github", OK: false, Message: "GitHub API base URL 未配置"}
	}
	token := strings.TrimSpace(request.Token)
	if token == "" {
		token = ResolvedGitHubToken()
	}
	if token == "" {
		return ConnectionCheck{Target: "github", OK: false, Message: "GitHub token 未配置，公共 API 仍可用但会很快限流"}
	}
	return ConnectionCheck{Target: "github", OK: true, Message: "GitHub base URL 与 token 已配置，当前检测使用表单中的 base URL"}
}

func CheckLLMConnection(providerName string) ConnectionCheck {
	cfg, _, _ := loadAppConfig()
	secrets, _, _ := loadSecretConfig()
	providerName = normalizeProviderName(providerName)
	for _, provider := range cfg.Providers {
		if provider.Name != providerName {
			continue
		}
		if !provider.Enabled {
			return ConnectionCheck{Target: providerName, OK: false, Message: "Provider 未启用"}
		}
		if strings.TrimSpace(provider.BaseURL) == "" {
			return ConnectionCheck{Target: providerName, OK: false, Message: "Base URL 未配置"}
		}
		if !providerTokenConfigured(providerName, secrets) {
			return ConnectionCheck{Target: providerName, OK: false, Message: "API token 未配置"}
		}
		return probeProvider(provider, resolvedProviderToken(providerName, secrets))
	}
	return ConnectionCheck{Target: providerName, OK: false, Message: "未知 provider"}
}

func CheckLLMConnectionRequest(request ProviderConnectionRequest) ConnectionCheck {
	providerName := normalizeProviderName(request.Provider)
	if providerName == "" {
		return ConnectionCheck{Target: "llm", OK: false, Message: "Provider 未配置"}
	}
	if !request.Enabled {
		return ConnectionCheck{Target: providerName, OK: false, Message: "Provider 未启用"}
	}
	if strings.TrimSpace(request.BaseURL) == "" {
		return ConnectionCheck{Target: providerName, OK: false, Message: "Base URL 未配置"}
	}
	token := strings.TrimSpace(request.Token)
	if token == "" {
		secrets, _, _ := loadSecretConfig()
		token = strings.TrimSpace(resolvedProviderToken(providerName, secrets))
	}
	if token == "" {
		return ConnectionCheck{Target: providerName, OK: false, Message: "API token 未配置"}
	}
	provider := ProviderConfig{
		Name:    providerName,
		Model:   strings.TrimSpace(request.Model),
		BaseURL: strings.TrimRight(strings.TrimSpace(request.BaseURL), "/"),
		Enabled: request.Enabled,
	}
	if provider.Model == "" {
		provider.Model = defaultModel(providerName)
	}
	return probeProvider(provider, token)
}

func ActiveRuntimeProvider() (RuntimeProvider, bool) {
	cfg, _, _ := loadAppConfig()
	secrets, _, _ := loadSecretConfig()
	active := normalizeProviderName(envOrValue("MODEL_PROVIDER", cfg.ModelProvider))
	if active == "" {
		active = "openai"
	}
	for _, provider := range cfg.Providers {
		if provider.Name != active || !provider.Enabled {
			continue
		}
		if provider.Model == "" {
			provider.Model = envOrValue("MODEL_NAME", defaultModel(active))
		}
		token := resolvedProviderToken(active, secrets)
		if strings.TrimSpace(provider.BaseURL) == "" || strings.TrimSpace(token) == "" {
			return RuntimeProvider{}, false
		}
		return RuntimeProvider{Provider: provider, Token: strings.TrimSpace(token)}, true
	}
	return RuntimeProvider{}, false
}

func probeProvider(provider ProviderConfig, token string) ConnectionCheck {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	switch normalizeProviderName(provider.Name) {
	case "anthropic":
		return probeAnthropic(ctx, provider, token)
	default:
		return probeOpenAICompatible(ctx, provider, token)
	}
}

func probeOpenAICompatible(ctx context.Context, provider ProviderConfig, token string) ConnectionCheck {
	endpoint := strings.TrimRight(provider.BaseURL, "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ConnectionCheck{Target: provider.Name, OK: false, Message: err.Error()}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ConnectionCheck{Target: provider.Name, OK: false, Message: "连接失败：" + err.Error()}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return ConnectionCheck{Target: provider.Name, OK: true, Message: "连接成功：OpenAI-compatible /models 可访问"}
	}
	return ConnectionCheck{Target: provider.Name, OK: false, Message: fmt.Sprintf("连接失败：HTTP %d %s", resp.StatusCode, summarizeBody(body))}
}

func probeAnthropic(ctx context.Context, provider ProviderConfig, token string) ConnectionCheck {
	endpoint := strings.TrimRight(provider.BaseURL, "/") + "/v1/messages"
	payload := map[string]any{
		"model":      provider.Model,
		"max_tokens": 1,
		"messages": []map[string]string{
			{"role": "user", "content": "ping"},
		},
	}
	data, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return ConnectionCheck{Target: provider.Name, OK: false, Message: err.Error()}
	}
	req.Header.Set("x-api-key", token)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ConnectionCheck{Target: provider.Name, OK: false, Message: "连接失败：" + err.Error()}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return ConnectionCheck{Target: provider.Name, OK: true, Message: "连接成功：Anthropic messages API 可访问"}
	}
	return ConnectionCheck{Target: provider.Name, OK: false, Message: fmt.Sprintf("连接失败：HTTP %d %s", resp.StatusCode, summarizeBody(body))}
}

func summarizeBody(body []byte) string {
	text := strings.TrimSpace(string(body))
	if text == "" {
		return ""
	}
	text = strings.ReplaceAll(text, "\n", " ")
	if len(text) > 220 {
		return text[:220] + "..."
	}
	return text
}

func defaultModel(provider string) string {
	switch provider {
	case "anthropic", "claude":
		return "claude-3-5-sonnet-latest"
	case "deepseek":
		return "deepseek-chat"
	default:
		return "gpt-4.1-mini"
	}
}

func providerTokenConfigured(provider string, secrets SecretConfig) bool {
	return strings.TrimSpace(resolvedProviderToken(provider, secrets)) != ""
}

func resolvedProviderToken(provider string, secrets SecretConfig) string {
	provider = normalizeProviderName(provider)
	switch provider {
	case "openai":
		if value := strings.TrimSpace(os.Getenv("OPENAI_API_KEY")); value != "" {
			return value
		}
		if secrets.ProviderTokens != nil && strings.TrimSpace(secrets.ProviderTokens["openai"]) != "" {
			return secrets.ProviderTokens["openai"]
		}
		return secrets.OpenAIKey
	case "anthropic":
		if value := strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")); value != "" {
			return value
		}
		if secrets.ProviderTokens != nil && strings.TrimSpace(secrets.ProviderTokens["anthropic"]) != "" {
			return secrets.ProviderTokens["anthropic"]
		}
		return secrets.AnthropicKey
	case "deepseek":
		if value := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")); value != "" {
			return value
		}
		if secrets.ProviderTokens != nil && strings.TrimSpace(secrets.ProviderTokens["deepseek"]) != "" {
			return secrets.ProviderTokens["deepseek"]
		}
		return secrets.DeepSeekKey
	default:
		if secrets.ProviderTokens != nil {
			return secrets.ProviderTokens[provider]
		}
		return ""
	}
}

func providerTokenMap(secrets SecretConfig) map[string]bool {
	configured := map[string]bool{}
	for _, provider := range []string{"openai", "anthropic", "deepseek", "custom"} {
		configured[provider] = providerTokenConfigured(provider, secrets)
	}
	return configured
}

func toLLMProviders(providers []ProviderConfig) []llm.ConfiguredProvider {
	result := make([]llm.ConfiguredProvider, 0, len(providers))
	for _, provider := range providers {
		result = append(result, llm.ConfiguredProvider{
			Name:    provider.Name,
			Model:   provider.Model,
			BaseURL: provider.BaseURL,
			Enabled: provider.Enabled,
		})
	}
	return result
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envOrValue(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envOrFile(envKey, fileValue string) bool {
	if strings.TrimSpace(os.Getenv(envKey)) != "" {
		return true
	}
	return strings.TrimSpace(fileValue) != ""
}

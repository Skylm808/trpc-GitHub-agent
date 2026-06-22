package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const appDirName = "trpc-GitHub-agent"

type AppConfig struct {
	ModelProvider      string           `yaml:"model_provider" json:"model_provider"`
	ModelName          string           `yaml:"model_name" json:"model_name"`
	UILanguage         string           `yaml:"ui_language" json:"ui_language"`
	GitHubBaseURL      string           `yaml:"github_base_url" json:"github_base_url"`
	Providers          []ProviderConfig `yaml:"providers" json:"providers"`
	DefaultLanguage    string           `yaml:"default_language" json:"default_language"`
	InputLanguage      string           `yaml:"input_language" json:"input_language"`
	DefaultDirection   string           `yaml:"default_direction" json:"default_direction"`
	DefaultDifficulty  string           `yaml:"default_difficulty" json:"default_difficulty"`
	DefaultMinStars    int              `yaml:"default_min_stars" json:"default_min_stars"`
	DefaultMaxStars    int              `yaml:"default_max_stars" json:"default_max_stars"`
	DefaultPushedAfter string           `yaml:"default_pushed_after" json:"default_pushed_after"`
	Theme              string           `yaml:"theme" json:"theme"`
}

type ProviderConfig struct {
	Name    string `yaml:"name" json:"name"`
	Model   string `yaml:"model" json:"model"`
	BaseURL string `yaml:"base_url" json:"base_url"`
	Enabled bool   `yaml:"enabled" json:"enabled"`
}

type SecretConfig struct {
	GitHubToken    string            `json:"github_token,omitempty"`
	ProviderTokens map[string]string `json:"provider_tokens,omitempty"`
	OpenAIKey      string            `json:"openai_api_key,omitempty"`
	AnthropicKey   string            `json:"anthropic_api_key,omitempty"`
	DeepSeekKey    string            `json:"deepseek_api_key,omitempty"`
}

type SettingsBundle struct {
	Config      AppConfig      `json:"config"`
	Status      SettingsStatus `json:"status"`
	ConfigPath  string         `json:"config_path"`
	SecretsPath string         `json:"secrets_path"`
}

type SettingsUpdate struct {
	Config         AppConfig         `json:"config"`
	GitHubToken    string            `json:"github_token"`
	ProviderTokens map[string]string `json:"provider_tokens"`
	OpenAIKey      string            `json:"openai_api_key"`
	AnthropicKey   string            `json:"anthropic_api_key"`
	DeepSeekKey    string            `json:"deepseek_api_key"`
}

func LoadSettingsBundle() SettingsBundle {
	cfg, configPath, _ := loadAppConfig()
	secrets, secretsPath, _ := loadSecretConfig()
	status := loadSettingsStatus(cfg, secrets)
	return SettingsBundle{
		Config:      cfg,
		Status:      status,
		ConfigPath:  configPath,
		SecretsPath: secretsPath,
	}
}

func SaveSettings(update SettingsUpdate) (SettingsBundle, error) {
	cfg := normalizeConfig(update.Config)
	configPath, err := configFilePath()
	if err != nil {
		return SettingsBundle{}, err
	}
	if err := writeYAML(configPath, cfg); err != nil {
		return SettingsBundle{}, err
	}

	secrets, _, err := loadSecretConfig()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return SettingsBundle{}, err
	}
	if update.GitHubToken != "" {
		secrets.GitHubToken = update.GitHubToken
	}
	if secrets.ProviderTokens == nil {
		secrets.ProviderTokens = map[string]string{}
	}
	for provider, token := range update.ProviderTokens {
		provider = normalizeProviderName(provider)
		if provider != "" && token != "" {
			secrets.ProviderTokens[provider] = token
		}
	}
	if update.OpenAIKey != "" {
		secrets.OpenAIKey = update.OpenAIKey
		secrets.ProviderTokens["openai"] = update.OpenAIKey
	}
	if update.AnthropicKey != "" {
		secrets.AnthropicKey = update.AnthropicKey
		secrets.ProviderTokens["anthropic"] = update.AnthropicKey
	}
	if update.DeepSeekKey != "" {
		secrets.DeepSeekKey = update.DeepSeekKey
		secrets.ProviderTokens["deepseek"] = update.DeepSeekKey
	}
	secretPath, err := secretFilePath()
	if err != nil {
		return SettingsBundle{}, err
	}
	if err := writeJSON(secretPath, secrets); err != nil {
		return SettingsBundle{}, err
	}
	return LoadSettingsBundle(), nil
}

func loadAppConfig() (AppConfig, string, error) {
	path, err := configFilePath()
	if err != nil {
		return normalizeConfig(AppConfig{}), "", err
	}
	cfg := normalizeConfig(AppConfig{})
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, path, nil
		}
		return cfg, path, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, path, fmt.Errorf("decode config yaml: %w", err)
	}
	cfg = normalizeConfig(cfg)
	return cfg, path, nil
}

func loadSecretConfig() (SecretConfig, string, error) {
	path, err := secretFilePath()
	if err != nil {
		return SecretConfig{}, "", err
	}
	var secrets SecretConfig
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return secrets, path, nil
		}
		return secrets, path, err
	}
	if err := json.Unmarshal(data, &secrets); err != nil {
		return secrets, path, fmt.Errorf("decode secrets json: %w", err)
	}
	return secrets, path, nil
}

func writeYAML(path string, cfg AppConfig) error {
	if err := ensureDir(filepath.Dir(path)); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func writeJSON(path string, secrets SecretConfig) error {
	if err := ensureDir(filepath.Dir(path)); err != nil {
		return err
	}
	data, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

func normalizeConfig(cfg AppConfig) AppConfig {
	if cfg.ModelProvider == "" {
		cfg.ModelProvider = "openai"
	}
	cfg.ModelProvider = normalizeProviderName(cfg.ModelProvider)
	if cfg.ModelName == "" {
		cfg.ModelName = defaultModel(cfg.ModelProvider)
	}
	if cfg.UILanguage == "" {
		cfg.UILanguage = "zh"
	}
	if cfg.GitHubBaseURL == "" {
		cfg.GitHubBaseURL = "https://api.github.com"
	}
	cfg.Providers = normalizeProviders(cfg.ModelProvider, cfg.ModelName, cfg.Providers)
	if cfg.DefaultLanguage == "" {
		cfg.DefaultLanguage = "Go"
	}
	if cfg.InputLanguage == "" {
		cfg.InputLanguage = "auto"
	}
	if cfg.DefaultDirection == "" {
		cfg.DefaultDirection = "agent"
	}
	if cfg.DefaultDifficulty == "" {
		cfg.DefaultDifficulty = "intermediate"
	}
	if cfg.DefaultMinStars <= 0 {
		cfg.DefaultMinStars = 100
	}
	if cfg.DefaultMaxStars <= 0 {
		cfg.DefaultMaxStars = 50000
	}
	if cfg.DefaultPushedAfter == "" {
		cfg.DefaultPushedAfter = "2025-01-01"
	}
	if cfg.Theme == "" {
		cfg.Theme = "dark"
	}
	return cfg
}

func normalizeProviders(activeProvider, activeModel string, providers []ProviderConfig) []ProviderConfig {
	byName := map[string]ProviderConfig{}
	for _, provider := range providers {
		name := normalizeProviderName(provider.Name)
		if name == "" {
			continue
		}
		provider.Name = name
		provider.BaseURL = strings.TrimRight(provider.BaseURL, "/")
		byName[name] = provider
	}
	defaults := []ProviderConfig{
		{Name: "openai", Model: "gpt-4.1-mini", BaseURL: "https://api.openai.com/v1", Enabled: true},
		{Name: "anthropic", Model: "claude-3-5-sonnet-latest", BaseURL: "https://api.anthropic.com", Enabled: true},
		{Name: "deepseek", Model: "deepseek-chat", BaseURL: "https://api.deepseek.com/v1", Enabled: true},
		{Name: "custom", Model: "gpt-4.1-mini", BaseURL: "", Enabled: false},
	}
	normalized := make([]ProviderConfig, 0, len(defaults))
	for _, fallback := range defaults {
		if existing, ok := byName[fallback.Name]; ok {
			if existing.Model == "" {
				existing.Model = fallback.Model
			}
			if existing.BaseURL == "" {
				existing.BaseURL = fallback.BaseURL
			}
			if existing.Name == activeProvider && activeModel != "" {
				existing.Model = activeModel
			}
			normalized = append(normalized, existing)
			continue
		}
		if fallback.Name == activeProvider && activeModel != "" {
			fallback.Model = activeModel
		}
		normalized = append(normalized, fallback)
	}
	return normalized
}

func normalizeProviderName(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "claude":
		return "anthropic"
	default:
		return strings.ToLower(strings.TrimSpace(provider))
	}
}

func configFilePath() (string, error) {
	root, err := baseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "config.yaml"), nil
}

func secretFilePath() (string, error) {
	root, err := baseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "secrets.json"), nil
}

func baseDir() (string, error) {
	if root := os.Getenv("TRPC_GITHUB_AGENT_HOME"); root != "" {
		return filepath.Join(root, appDirName), nil
	}
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, appDirName), nil
}

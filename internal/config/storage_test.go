package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveSettingsPersistsYamlAndSecrets(t *testing.T) {
	root := t.TempDir()
	t.Setenv("TRPC_GITHUB_AGENT_HOME", root)
	t.Setenv("MODEL_PROVIDER", "")
	t.Setenv("MODEL_NAME", "")
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("DEEPSEEK_API_KEY", "")

	_, err := SaveSettings(SettingsUpdate{
		Config: AppConfig{
			ModelProvider: "anthropic",
			ModelName:     "claude-sonnet-4",
			UILanguage:    "zh",
			GitHubBaseURL: "https://github-proxy.example.com",
			Providers: []ProviderConfig{
				{Name: "anthropic", Model: "claude-sonnet-4", BaseURL: "https://claude-relay.example.com", Enabled: true},
				{Name: "deepseek", Model: "deepseek-reasoner", BaseURL: "https://deepseek.example.com/v1", Enabled: true},
			},
			DefaultLanguage:    "Go",
			InputLanguage:      "zh",
			DefaultDirection:   "agent",
			DefaultDifficulty:  "advanced",
			DefaultMinStars:    200,
			DefaultMaxStars:    20000,
			DefaultPushedAfter: "2025-05-01",
			Theme:              "dark",
		},
		GitHubToken: "github-token",
		ProviderTokens: map[string]string{
			"openai":    "openai-token",
			"anthropic": "anthropic-token",
			"deepseek":  "deepseek-token",
		},
	})
	if err != nil {
		t.Fatalf("save settings: %v", err)
	}

	configPath := filepath.Join(root, appDirName, "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("expected config file: %v", err)
	}
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config file: %v", err)
	}
	if strings.Contains(string(configData), "github-token") || strings.Contains(string(configData), "openai-token") {
		t.Fatal("config.yaml must not contain tokens")
	}
	if !strings.Contains(string(configData), "github_base_url") || !strings.Contains(string(configData), "providers:") {
		t.Fatalf("expected base urls and providers in config.yaml: %s", string(configData))
	}
	secretPath := filepath.Join(root, appDirName, "secrets.json")
	if _, err := os.Stat(secretPath); err != nil {
		t.Fatalf("expected secrets file: %v", err)
	}

	bundle := LoadSettingsBundle()
	if bundle.Config.ModelProvider != "anthropic" || bundle.Config.DefaultMaxStars != 20000 || bundle.Config.GitHubBaseURL != "https://github-proxy.example.com" {
		t.Fatalf("unexpected config bundle: %#v", bundle.Config)
	}
	if !bundle.Status.GitHubTokenConfigured || !bundle.Status.AnthropicConfigured {
		t.Fatalf("expected secret status to be configured: %#v", bundle.Status)
	}
}

package config

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoadSettingsStatusReadsProviderKeys(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "github-token")
	t.Setenv("MODEL_PROVIDER", "deepseek")
	t.Setenv("MODEL_NAME", "deepseek-reasoner")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("DEEPSEEK_API_KEY", "deepseek-token")

	status := LoadSettingsStatus()
	if !status.GitHubTokenConfigured {
		t.Fatal("expected GitHub token to be configured")
	}
	if status.ModelProvider != "deepseek" || status.ModelName != "deepseek-reasoner" {
		t.Fatalf("unexpected model selection: %#v", status)
	}
	if !status.DeepSeekConfigured || !status.ActiveProviderReady {
		t.Fatalf("expected active DeepSeek provider to be ready: %#v", status)
	}
	if !status.DeterministicModeReady {
		t.Fatal("deterministic mode should not depend on LLM keys")
	}
	if len(status.Providers) != 4 {
		t.Fatalf("expected provider list, got %#v", status.Providers)
	}
}

func TestLoadSettingsStatusDefaultsToOpenAI(t *testing.T) {
	t.Setenv("TRPC_GITHUB_AGENT_HOME", t.TempDir())
	t.Setenv("MODEL_PROVIDER", "")
	t.Setenv("MODEL_NAME", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("DEEPSEEK_API_KEY", "")

	status := LoadSettingsStatus()
	if status.ModelProvider != "openai" {
		t.Fatalf("expected openai default provider, got %q", status.ModelProvider)
	}
	if status.ActiveProviderReady {
		t.Fatal("expected active provider to require user-supplied key")
	}
	if !status.DeterministicModeReady {
		t.Fatal("expected deterministic mode to remain ready")
	}
	if len(status.Providers) != 4 {
		t.Fatalf("expected provider list, got %#v", status.Providers)
	}
}

func TestCheckLLMConnectionProbesOpenAICompatibleEndpoint(t *testing.T) {
	root := t.TempDir()
	t.Setenv("TRPC_GITHUB_AGENT_HOME", root)
	t.Setenv("OPENAI_API_KEY", "")
	var sawAuth bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Fatalf("unexpected probe path %s", r.URL.Path)
		}
		sawAuth = r.Header.Get("Authorization") == "Bearer local-token"
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	_, err := SaveSettings(SettingsUpdate{
		Config: AppConfig{
			ModelProvider: "custom",
			ModelName:     "relay-model",
			Providers: []ProviderConfig{
				{Name: "custom", Model: "relay-model", BaseURL: server.URL + "/v1", Enabled: true},
			},
		},
		ProviderTokens: map[string]string{"custom": "local-token"},
	})
	if err != nil {
		t.Fatalf("save settings: %v", err)
	}

	check := CheckLLMConnection("custom")
	if !check.OK {
		t.Fatalf("expected successful probe, got %#v", check)
	}
	if !sawAuth {
		t.Fatal("expected probe to send bearer token")
	}
}

func TestCheckLLMConnectionRequestUsesSavedTokenWhenDraftTokenEmpty(t *testing.T) {
	root := t.TempDir()
	t.Setenv("TRPC_GITHUB_AGENT_HOME", root)
	t.Setenv("OPENAI_API_KEY", "")
	var sawAuth bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization") == "Bearer saved-token"
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	_, err := SaveSettings(SettingsUpdate{
		Config: AppConfig{
			ModelProvider: "custom",
			Providers: []ProviderConfig{
				{Name: "custom", Model: "relay-model", BaseURL: "https://saved.example.com/v1", Enabled: true},
			},
		},
		ProviderTokens: map[string]string{"custom": "saved-token"},
	})
	if err != nil {
		t.Fatalf("save settings: %v", err)
	}

	check := CheckLLMConnectionRequest(ProviderConnectionRequest{
		Provider: "custom",
		Model:    "relay-model",
		BaseURL:  server.URL,
		Enabled:  true,
	})
	if !check.OK {
		t.Fatalf("expected draft probe to use saved token, got %#v", check)
	}
	if !sawAuth {
		t.Fatal("expected saved bearer token")
	}
}

func TestCheckGitHubConnectionRequestUsesDraftBaseURLAndSavedToken(t *testing.T) {
	root := t.TempDir()
	t.Setenv("TRPC_GITHUB_AGENT_HOME", root)
	t.Setenv("GITHUB_TOKEN", "")
	_, err := SaveSettings(SettingsUpdate{
		Config:      AppConfig{GitHubBaseURL: "https://saved.example.com"},
		GitHubToken: "saved-github-token",
	})
	if err != nil {
		t.Fatalf("save settings: %v", err)
	}

	check := CheckGitHubConnectionRequest(GitHubConnectionRequest{BaseURL: "https://draft.example.com"})
	if !check.OK {
		t.Fatalf("expected draft GitHub check to use saved token, got %#v", check)
	}
	if check.Target != "github" {
		t.Fatalf("unexpected target %q", check.Target)
	}
}

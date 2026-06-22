package config

import "testing"

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
}

func TestLoadSettingsStatusDefaultsToOpenAI(t *testing.T) {
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
}

package llmchat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientChatCallsOpenAICompatibleEndpoint(t *testing.T) {
	var sawAuth bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		sawAuth = r.Header.Get("Authorization") == "Bearer test-token"
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"回答"}}]}`))
	}))
	defer server.Close()

	client := Client{
		Provider: "custom",
		Model:    "test-model",
		BaseURL:  server.URL + "/v1",
		Token:    "test-token",
	}
	answer, err := client.Chat(context.Background(), []Message{{Role: "user", Content: "ping"}})
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if answer != "回答" {
		t.Fatalf("unexpected answer %q", answer)
	}
	if !sawAuth {
		t.Fatal("expected bearer auth header")
	}
}

func TestClientChatCallsAnthropicMessagesEndpoint(t *testing.T) {
	var sawAPIKey bool
	var sawVersion bool
	var body struct {
		Model    string    `json:"model"`
		System   string    `json:"system"`
		Messages []Message `json:"messages"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		sawAPIKey = r.Header.Get("x-api-key") == "anthropic-token"
		sawVersion = r.Header.Get("anthropic-version") == "2023-06-01"
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"content":[{"type":"text","text":"Claude 回答"}]}`))
	}))
	defer server.Close()

	client := Client{
		Provider: "anthropic",
		Model:    "claude-sonnet-test",
		BaseURL:  server.URL,
		Token:    "anthropic-token",
	}
	answer, err := client.Chat(context.Background(), []Message{
		{Role: "system", Content: "系统提示"},
		{Role: "user", Content: "ping"},
	})
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if answer != "Claude 回答" {
		t.Fatalf("unexpected answer %q", answer)
	}
	if !sawAPIKey || !sawVersion {
		t.Fatalf("expected anthropic headers, apiKey=%v version=%v", sawAPIKey, sawVersion)
	}
	if body.Model != "claude-sonnet-test" || body.System != "系统提示" {
		t.Fatalf("unexpected request body: %#v", body)
	}
	if len(body.Messages) != 1 || body.Messages[0].Role != "user" || body.Messages[0].Content != "ping" {
		t.Fatalf("unexpected anthropic messages: %#v", body.Messages)
	}
}

func TestSplitAnthropicMessagesNormalizesUnsupportedRoles(t *testing.T) {
	system, messages := splitAnthropicMessages([]Message{
		{Role: "system", Content: "a"},
		{Role: "tool", Content: "b"},
		{Role: "assistant", Content: "c"},
	})
	if system != "a" {
		t.Fatalf("unexpected system %q", system)
	}
	if len(messages) != 2 || messages[0].Role != "user" || messages[1].Role != "assistant" {
		t.Fatalf("unexpected messages %#v", messages)
	}
}

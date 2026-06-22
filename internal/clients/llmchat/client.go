package llmchat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	Provider   string
	Model      string
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (c Client) Chat(ctx context.Context, messages []Message) (string, error) {
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
	if strings.EqualFold(c.Provider, "anthropic") || strings.EqualFold(c.Provider, "claude") {
		return c.chatAnthropic(ctx, messages)
	}
	return c.chatOpenAICompatible(ctx, messages)
}

func (c Client) chatOpenAICompatible(ctx context.Context, messages []Message) (string, error) {
	endpoint := strings.TrimRight(c.BaseURL, "/") + "/chat/completions"
	payload := map[string]any{
		"model":       c.Model,
		"messages":    messages,
		"temperature": 0.2,
		"max_tokens":  900,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("llm request failed with HTTP %d: %s", resp.StatusCode, trimBody(body))
	}
	var parsed struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("decode llm response: %w", err)
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("llm returned an empty answer")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

func (c Client) chatAnthropic(ctx context.Context, messages []Message) (string, error) {
	endpoint := strings.TrimRight(c.BaseURL, "/") + "/v1/messages"
	system, anthropicMessages := splitAnthropicMessages(messages)
	payload := map[string]any{
		"model":       c.Model,
		"messages":    anthropicMessages,
		"temperature": 0.2,
		"max_tokens":  900,
	}
	if system != "" {
		payload["system"] = system
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("x-api-key", c.Token)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("llm request failed with HTTP %d: %s", resp.StatusCode, trimBody(body))
	}
	var parsed struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("decode anthropic response: %w", err)
	}
	var parts []string
	for _, item := range parsed.Content {
		if item.Type == "text" && strings.TrimSpace(item.Text) != "" {
			parts = append(parts, strings.TrimSpace(item.Text))
		}
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("llm returned an empty answer")
	}
	return strings.Join(parts, "\n"), nil
}

func splitAnthropicMessages(messages []Message) (string, []Message) {
	var systems []string
	var result []Message
	for _, message := range messages {
		role := strings.ToLower(strings.TrimSpace(message.Role))
		content := strings.TrimSpace(message.Content)
		if content == "" {
			continue
		}
		switch role {
		case "system":
			systems = append(systems, content)
		case "assistant":
			result = append(result, Message{Role: "assistant", Content: content})
		default:
			result = append(result, Message{Role: "user", Content: content})
		}
	}
	if len(result) == 0 {
		result = append(result, Message{Role: "user", Content: "ping"})
	}
	return strings.Join(systems, "\n\n"), result
}

func trimBody(body []byte) string {
	text := strings.TrimSpace(string(body))
	text = strings.ReplaceAll(text, "\n", " ")
	if len(text) > 300 {
		return text[:300] + "..."
	}
	return text
}

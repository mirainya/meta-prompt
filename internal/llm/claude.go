package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ClaudeProvider struct {
	apiKey    string
	baseURL   string
	model     string
	maxTokens int
}

func NewClaudeProvider(apiKey, model string, maxTokens int) *ClaudeProvider {
	return &ClaudeProvider{apiKey: apiKey, baseURL: "https://api.anthropic.com", model: model, maxTokens: maxTokens}
}

func NewClaudeProviderWithBase(apiKey, baseURL, model string, maxTokens int) *ClaudeProvider {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &ClaudeProvider{apiKey: apiKey, baseURL: baseURL, model: model, maxTokens: maxTokens}
}

func (p *ClaudeProvider) Name() string { return "claude" }

func (p *ClaudeProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	maxTokens := p.maxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}

	var system string
	var messages []map[string]string
	for _, m := range req.Messages {
		if m.Role == "system" {
			system = m.Content
			continue
		}
		messages = append(messages, map[string]string{"role": m.Role, "content": m.Content})
	}

	body := map[string]any{
		"model":      p.model,
		"max_tokens": maxTokens,
		"messages":   messages,
	}
	if system != "" {
		body["system"] = system
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}

	data, _ := json.Marshal(body)
	url := p.baseURL + "/v1/messages"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("claude api error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	if len(result.Content) == 0 {
		return nil, fmt.Errorf("claude returned empty content")
	}

	return &ChatResponse{Content: result.Content[0].Text}, nil
}

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

type OpenAIProvider struct {
	apiKey    string
	baseURL   string
	model     string
	maxTokens int
}

func NewOpenAIProvider(apiKey, model string, maxTokens int) *OpenAIProvider {
	return &OpenAIProvider{apiKey: apiKey, baseURL: "https://api.openai.com", model: model, maxTokens: maxTokens}
}

func NewOpenAIProviderWithBase(apiKey, baseURL, model string, maxTokens int) *OpenAIProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &OpenAIProvider{apiKey: apiKey, baseURL: baseURL, model: model, maxTokens: maxTokens}
}

func (p *OpenAIProvider) Name() string { return "openai" }

func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	maxTokens := p.maxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}

	var messages []map[string]string
	for _, m := range req.Messages {
		messages = append(messages, map[string]string{"role": m.Role, "content": m.Content})
	}

	body := map[string]any{
		"model":      p.model,
		"max_tokens": maxTokens,
		"messages":   messages,
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	if req.JSONMode {
		body["response_format"] = map[string]string{"type": "json_object"}
	}

	data, _ := json.Marshal(body)
	url := p.baseURL + "/v1/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai api error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("openai returned empty choices")
	}

	return &ChatResponse{Content: result.Choices[0].Message.Content}, nil
}

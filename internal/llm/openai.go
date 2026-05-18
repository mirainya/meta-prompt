package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type OpenAIProvider struct {
	apiKey   string
	baseURL  string
	model    string
	proxyURL string
}

func (p *OpenAIProvider) Name() string { return p.model }

func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
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
	apiURL := strings.TrimRight(p.baseURL, "/") + "/v1/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", p.apiKey)

	client := p.httpClient()
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(respBody))
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
		return nil, fmt.Errorf("api returned empty choices")
	}

	return &ChatResponse{Content: result.Choices[0].Message.Content}, nil
}

func (p *OpenAIProvider) httpClient() *http.Client {
	if p.proxyURL == "" {
		return http.DefaultClient
	}
	u, err := url.Parse(p.proxyURL)
	if err != nil {
		return http.DefaultClient
	}
	return &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(u)},
	}
}

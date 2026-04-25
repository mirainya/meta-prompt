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

type GeminiProvider struct {
	apiKey    string
	baseURL   string
	model     string
	maxTokens int
}

func NewGeminiProvider(apiKey, model string, maxTokens int) *GeminiProvider {
	return &GeminiProvider{apiKey: apiKey, baseURL: "https://generativelanguage.googleapis.com", model: model, maxTokens: maxTokens}
}

func NewGeminiProviderWithBase(apiKey, baseURL, model string, maxTokens int) *GeminiProvider {
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &GeminiProvider{apiKey: apiKey, baseURL: baseURL, model: model, maxTokens: maxTokens}
}

func (p *GeminiProvider) Name() string { return "gemini" }

func (p *GeminiProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	maxTokens := p.maxTokens
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}

	var systemInstruction string
	var contents []map[string]any
	for _, m := range req.Messages {
		if m.Role == "system" {
			systemInstruction = m.Content
			continue
		}
		role := m.Role
		if role == "assistant" {
			role = "model"
		}
		contents = append(contents, map[string]any{
			"role":  role,
			"parts": []map[string]string{{"text": m.Content}},
		})
	}

	body := map[string]any{
		"contents": contents,
		"generationConfig": map[string]any{
			"maxOutputTokens": maxTokens,
		},
	}
	if systemInstruction != "" {
		body["systemInstruction"] = map[string]any{
			"parts": []map[string]string{{"text": systemInstruction}},
		}
	}
	if req.Temperature > 0 {
		body["generationConfig"].(map[string]any)["temperature"] = req.Temperature
	}
	if req.JSONMode {
		body["generationConfig"].(map[string]any)["responseMimeType"] = "application/json"
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", p.baseURL, p.model, p.apiKey)

	data, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini api error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned empty content")
	}

	return &ChatResponse{Content: result.Candidates[0].Content.Parts[0].Text}, nil
}

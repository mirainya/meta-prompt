package service

import (
	"context"
	"encoding/json"
	"fmt"

	"meta-prompt/internal/llm"
)

type Architect struct {
	providers *llm.ProviderManager
}

func NewArchitect(providers *llm.ProviderManager) *Architect {
	return &Architect{providers: providers}
}

func (a *Architect) Run(ctx context.Context, modelCode string, metaPrompt string, input string, analysis json.RawMessage) (json.RawMessage, error) {
	userContent := fmt.Sprintf("## 原始需求\n%s\n\n## 需求分析结果\n%s", input, string(analysis))

	resp, err := a.providers.Chat(ctx, modelCode, llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: metaPrompt},
			{Role: "user", Content: userContent},
		},
		JSONMode: true,
	})
	if err != nil {
		return nil, fmt.Errorf("architect call failed: %w", err)
	}

	raw := json.RawMessage(resp.Content)
	if !json.Valid(raw) {
		return nil, fmt.Errorf("architect returned invalid JSON")
	}
	return raw, nil
}

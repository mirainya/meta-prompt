package service

import (
	"context"
	"encoding/json"
	"fmt"

	"meta-prompt/internal/llm"
)

type Analyzer struct {
	providers *llm.ProviderManager
}

func NewAnalyzer(providers *llm.ProviderManager) *Analyzer {
	return &Analyzer{providers: providers}
}

func (a *Analyzer) Run(ctx context.Context, providerName string, metaPrompt string, input string) (json.RawMessage, error) {
	provider, ok := a.providers.Get(providerName)
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}

	resp, err := provider.Chat(ctx, llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: metaPrompt},
			{Role: "user", Content: input},
		},
		JSONMode: true,
	})
	if err != nil {
		return nil, fmt.Errorf("analyzer call failed: %w", err)
	}

	raw := json.RawMessage(resp.Content)
	if !json.Valid(raw) {
		return nil, fmt.Errorf("analyzer returned invalid JSON")
	}
	return raw, nil
}

package service

import (
	"context"
	"encoding/json"
	"fmt"

	"meta-prompt/internal/llm"
)

type Reviewer struct {
	providers *llm.ProviderManager
}

func NewReviewer(providers *llm.ProviderManager) *Reviewer {
	return &Reviewer{providers: providers}
}

// ReviewOne 审核单组提示词
func (r *Reviewer) ReviewOne(ctx context.Context, providerName string, metaPrompt string, input string, blueprint json.RawMessage, currentPrompt json.RawMessage, currentOrder int, totalCount int, previousReviewed []json.RawMessage) (json.RawMessage, error) {
	provider, ok := r.providers.Get(providerName)
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}

	userContent := fmt.Sprintf("## 原始需求\n%s\n\n## 架构师蓝图\n%s\n\n## 当前审核的提示词（第%d组，共%d组）\n%s",
		input, string(blueprint), currentOrder, totalCount, string(currentPrompt))

	if len(previousReviewed) > 0 {
		userContent += "\n\n## 已审核通过的前序组提示词"
		for i, p := range previousReviewed {
			userContent += fmt.Sprintf("\n\n### 第%d组\n%s", i+1, string(p))
		}
	}

	resp, err := provider.Chat(ctx, llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: metaPrompt},
			{Role: "user", Content: userContent},
		},
		JSONMode: true,
	})
	if err != nil {
		return nil, fmt.Errorf("reviewer call failed: %w", err)
	}

	raw := json.RawMessage(resp.Content)
	if !json.Valid(raw) {
		return nil, fmt.Errorf("reviewer returned invalid JSON")
	}
	return raw, nil
}

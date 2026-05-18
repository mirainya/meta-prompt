package service

import (
	"context"
	"encoding/json"
	"fmt"

	"meta-prompt/internal/llm"
)

type Writer struct {
	providers *llm.ProviderManager
}

func NewWriter(providers *llm.ProviderManager) *Writer {
	return &Writer{providers: providers}
}

func (w *Writer) RunOne(ctx context.Context, modelCode string, metaPrompt string, input string, analysis json.RawMessage, blueprint json.RawMessage, currentGroup json.RawMessage, previousPrompts []json.RawMessage) (json.RawMessage, error) {
	userContent := fmt.Sprintf("## 原始需求\n%s\n\n## 需求分析结果\n%s\n\n## 完整工作流蓝图\n%s\n\n## 当前需要撰写的提示词蓝图\n%s",
		input, string(analysis), string(blueprint), string(currentGroup))

	if len(previousPrompts) > 0 {
		userContent += "\n\n## 前序组已完成的提示词文本"
		for i, p := range previousPrompts {
			userContent += fmt.Sprintf("\n\n### 第%d组\n%s", i+1, string(p))
		}
	}

	resp, err := w.providers.Chat(ctx, modelCode, llm.ChatRequest{
		Messages: []llm.Message{
			{Role: "system", Content: metaPrompt},
			{Role: "user", Content: userContent},
		},
		JSONMode: true,
	})
	if err != nil {
		return nil, fmt.Errorf("writer call failed: %w", err)
	}

	raw := json.RawMessage(resp.Content)
	if !json.Valid(raw) {
		return nil, fmt.Errorf("writer returned invalid JSON")
	}
	return raw, nil
}

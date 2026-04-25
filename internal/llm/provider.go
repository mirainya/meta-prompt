package llm

import "context"

type Provider interface {
	Name() string
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

type ChatRequest struct {
	Messages    []Message
	Temperature float64
	MaxTokens   int
	JSONMode    bool
}

type ChatResponse struct {
	Content string
}

type Message struct {
	Role    string
	Content string
}

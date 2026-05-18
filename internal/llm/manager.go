package llm

import (
	"context"
	"fmt"
	"strings"

	"meta-prompt/internal/model"
)

type ModelResolver interface {
	GetModelByCode(code string) (*model.ChannelModel, error)
}

type ProviderManager struct {
	resolver ModelResolver
}

func NewProviderManager(resolver ModelResolver) *ProviderManager {
	return &ProviderManager{resolver: resolver}
}

func (m *ProviderManager) GetProvider(modelCode string) (Provider, error) {
	cm, err := m.resolver.GetModelByCode(modelCode)
	if err != nil {
		return nil, fmt.Errorf("model not found: %s", modelCode)
	}
	if !cm.Enabled {
		return nil, fmt.Errorf("model disabled: %s", modelCode)
	}
	if !cm.Source.Enabled {
		return nil, fmt.Errorf("channel source disabled: %s", cm.Source.Name)
	}

	baseURL := strings.TrimRight(cm.Source.BaseURL, "/")
	p := &OpenAIProvider{
		apiKey:   cm.Source.APIKey,
		baseURL:  baseURL,
		model:    cm.ModelCode,
		proxyURL: cm.Source.ProxyURL,
	}
	return p, nil
}

func (m *ProviderManager) Chat(ctx context.Context, modelCode string, req ChatRequest) (*ChatResponse, error) {
	p, err := m.GetProvider(modelCode)
	if err != nil {
		return nil, err
	}
	return p.Chat(ctx, req)
}

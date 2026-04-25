package llm

import (
	"sync"
)

type ProviderManager struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

func NewProviderManager() *ProviderManager {
	return &ProviderManager{providers: make(map[string]Provider)}
}

func (m *ProviderManager) Set(name string, p Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[name] = p
}

func (m *ProviderManager) Remove(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.providers, name)
}

func (m *ProviderManager) Get(name string) (Provider, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.providers[name]
	return p, ok
}

func (m *ProviderManager) All() map[string]Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cp := make(map[string]Provider, len(m.providers))
	for k, v := range m.providers {
		cp[k] = v
	}
	return cp
}

package store

import (
	"meta-prompt/internal/model"

	"gorm.io/gorm"
)

type LLMConfigStore struct {
	db *gorm.DB
}

func NewLLMConfigStore(db *gorm.DB) *LLMConfigStore {
	return &LLMConfigStore{db: db}
}

func (s *LLMConfigStore) List() ([]model.LLMConfig, error) {
	var configs []model.LLMConfig
	err := s.db.Order("provider").Find(&configs).Error
	return configs, err
}

func (s *LLMConfigStore) ListEnabled() ([]model.LLMConfig, error) {
	var configs []model.LLMConfig
	err := s.db.Where("enabled = ? AND api_key != ''", true).Order("provider").Find(&configs).Error
	return configs, err
}

func (s *LLMConfigStore) GetByProvider(provider string) (*model.LLMConfig, error) {
	var cfg model.LLMConfig
	err := s.db.Where("provider = ?", provider).First(&cfg).Error
	return &cfg, err
}

func (s *LLMConfigStore) Upsert(cfg *model.LLMConfig) error {
	var existing model.LLMConfig
	err := s.db.Where("provider = ?", cfg.Provider).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return s.db.Create(cfg).Error
	}
	if err != nil {
		return err
	}
	return s.db.Model(&existing).Updates(map[string]any{
		"type":       cfg.Type,
		"api_key":    cfg.APIKey,
		"base_url":   cfg.BaseURL,
		"model":      cfg.Model,
		"max_tokens": cfg.MaxTokens,
		"enabled":    cfg.Enabled,
	}).Error
}

func (s *LLMConfigStore) Delete(provider string) error {
	return s.db.Where("provider = ?", provider).Delete(&model.LLMConfig{}).Error
}

package model

import "time"

// Type 取值: openai_compatible, claude, gemini
type LLMConfig struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Provider  string    `json:"provider" gorm:"size:50;uniqueIndex;not null"`
	Type      string    `json:"type" gorm:"size:30;default:openai_compatible;not null"`
	APIKey    string    `json:"api_key" gorm:"size:500"`
	BaseURL   string    `json:"base_url" gorm:"size:500"`
	Model     string    `json:"model" gorm:"size:100"`
	MaxTokens int       `json:"max_tokens" gorm:"default:4096"`
	Enabled   bool      `json:"enabled" gorm:"default:true"`
	UpdatedAt time.Time `json:"updated_at"`
}

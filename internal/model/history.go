package model

import (
	"encoding/json"
	"time"
)

type History struct {
	ID              int64           `json:"id" gorm:"primaryKey"`
	UserID          int64           `json:"user_id" gorm:"index;not null;default:0"`
	Input           string          `json:"input" gorm:"type:text;not null"`
	LLMProvider     string          `json:"llm_provider" gorm:"size:50;not null"`
	Status          string          `json:"status" gorm:"size:20;not null;default:'running';index"`
	CurrentStep     int             `json:"current_step" gorm:"not null;default:0"`
	ErrorMsg        string          `json:"error_msg,omitempty" gorm:"type:text"`
	ReasonerOutput  json.RawMessage `json:"reasoner_output" gorm:"type:jsonb"`
	ArchitectOutput json.RawMessage `json:"architect_output,omitempty" gorm:"type:jsonb"`
	GeneratorOutput json.RawMessage `json:"generator_output" gorm:"type:jsonb"`
	ReviewerOutput  json.RawMessage `json:"reviewer_output,omitempty" gorm:"type:jsonb"`
	TemplateIDs     json.RawMessage `json:"template_ids" gorm:"type:jsonb"`
	DurationMs      int             `json:"duration_ms" gorm:"not null"`
	WebhookURL      string          `json:"webhook_url,omitempty" gorm:"type:text"`
	WebhookSecret   string          `json:"-" gorm:"type:text"`
	Source          string          `json:"source" gorm:"size:20;not null;default:'web'"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

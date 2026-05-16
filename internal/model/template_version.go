package model

import "time"

type TemplateVersion struct {
	ID         int64     `json:"id" gorm:"primaryKey"`
	TemplateID int64     `json:"template_id" gorm:"index;not null"`
	Prompt     string    `json:"prompt" gorm:"type:text;not null"`
	Version    int       `json:"version" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at"`
}

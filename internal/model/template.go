package model

import "time"

type Template struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Description string    `json:"description"`
	Stage       string    `json:"stage" gorm:"size:20;not null"`
	Prompt      string    `json:"prompt" gorm:"type:text;not null"`
	Version     int       `json:"version" gorm:"default:1;not null"`
	IsDefault   bool      `json:"is_default" gorm:"default:false;not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

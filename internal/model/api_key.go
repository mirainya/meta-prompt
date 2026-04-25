package model

import "time"

type APIKey struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	KeyHash   string    `json:"-" gorm:"size:64;uniqueIndex;not null"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	IsActive  bool      `json:"is_active" gorm:"default:true;not null"`
	CreatedAt time.Time `json:"created_at"`
}

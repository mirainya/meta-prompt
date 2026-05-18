package model

import "time"

type APIKey struct {
	ID           int64      `json:"id" gorm:"primaryKey"`
	KeyHash      string     `json:"-" gorm:"size:64;uniqueIndex;not null"`
	RawKey       string     `json:"raw_key" gorm:"size:70"`
	Prefix       string     `json:"prefix" gorm:"size:12;not null"`
	Name         string     `json:"name" gorm:"size:100;not null"`
	UserID       int64      `json:"user_id" gorm:"index"`
	RateLimit    int        `json:"rate_limit" gorm:"default:60;not null"`
	CreditsQuota int        `json:"credits_quota" gorm:"default:-1;not null"`
	IsActive     bool       `json:"is_active" gorm:"default:true;not null"`
	LastUsedAt   *time.Time `json:"last_used_at"`
	CreatedAt    time.Time  `json:"created_at"`
}

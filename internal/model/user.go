package model

import "time"

type User struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"size:50;uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"size:100;not null"`
	Role         string    `json:"role" gorm:"size:20;default:user;not null"`
	Credits      int       `json:"credits" gorm:"default:10;not null"`
	Disabled     bool      `json:"disabled" gorm:"default:false;not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

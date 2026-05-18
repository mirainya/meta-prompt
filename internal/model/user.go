package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type StringArray []string

func (a *StringArray) Scan(val interface{}) error {
	if val == nil {
		*a = nil
		return nil
	}
	return json.Unmarshal(val.([]byte), a)
}

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return json.Marshal(a)
}

type User struct {
	ID            int64       `json:"id" gorm:"primaryKey"`
	Username      string      `json:"username" gorm:"size:50;uniqueIndex;not null"`
	PasswordHash  string      `json:"-" gorm:"size:100;not null"`
	Role          string      `json:"role" gorm:"size:20;default:user;not null"`
	Credits       int         `json:"credits" gorm:"default:10;not null"`
	Disabled      bool        `json:"disabled" gorm:"default:false;not null"`
	AllowedModels StringArray `json:"allowed_models" gorm:"type:jsonb"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

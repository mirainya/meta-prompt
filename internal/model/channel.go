package model

import "time"

type ChannelSource struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:100;uniqueIndex;not null"`
	BaseURL   string    `json:"base_url" gorm:"size:500;not null"`
	APIKey    string    `json:"api_key" gorm:"size:500;not null"`
	ProxyURL  string    `json:"proxy_url" gorm:"size:500"`
	Enabled   bool      `json:"enabled" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ChannelModel struct {
	ID               int64         `json:"id" gorm:"primaryKey"`
	SourceID         int64         `json:"source_id" gorm:"index;not null"`
	ModelCode        string        `json:"model_code" gorm:"size:100;uniqueIndex;not null"`
	ModelType        string        `json:"model_type" gorm:"size:20;not null"` // chat, image, video
	BillingType      string        `json:"billing_type" gorm:"size:20;default:per_call;not null"` // per_call, per_token
	CreditsPerCall   int           `json:"credits_per_call" gorm:"default:1;not null"`
	InputTokenPrice  int           `json:"input_token_price" gorm:"default:1;not null"`  // 每千输入token积分
	OutputTokenPrice int           `json:"output_token_price" gorm:"default:2;not null"` // 每千输出token积分
	Enabled          bool          `json:"enabled" gorm:"default:true"`
	IsDefault        bool          `json:"is_default" gorm:"default:false"`
	SyncedAt         time.Time     `json:"synced_at"`
	Source           ChannelSource `json:"-" gorm:"foreignKey:SourceID"`
}

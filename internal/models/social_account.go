package models

import (
	"time"
)

// SocialAccount represents the social_accounts table.
// Links OAuth2 providers (Google, Facebook, Twitter) to a user.
type SocialAccount struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null;index" json:"user_id"`
	Provider   string    `gorm:"size:50;not null" json:"provider"`
	ProviderID string    `gorm:"size:255;not null" json:"provider_id"`
	CreatedAt  time.Time `json:"created_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

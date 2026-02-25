package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Review represents the reviews table.
// Type: "place", "food", or "news"
// Status: "pending", "approved", or "rejected"
// Images stored as JSON array of URLs.
type Review struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Title     string         `gorm:"size:500;not null" json:"title"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	Type      string         `gorm:"size:20;not null" json:"type"`
	Status    string         `gorm:"size:20;default:'pending';not null" json:"status"`
	LikeCount int            `gorm:"default:0" json:"like_count"`
	Images    datatypes.JSON `gorm:"type:json" json:"images"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Comments []Comment `gorm:"foreignKey:ReviewID" json:"comments,omitempty"`
}

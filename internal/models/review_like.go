package models

import (
	"time"
)

// ReviewLike represents the review_likes junction table.
// Composite primary key: (UserID, ReviewID)
type ReviewLike struct {
	UserID    uint      `gorm:"primaryKey" json:"user_id"`
	ReviewID  uint      `gorm:"primaryKey" json:"review_id"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	User   *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Review *Review `gorm:"foreignKey:ReviewID" json:"review,omitempty"`
}

package models

import (
	"time"
)

// Rating represents the ratings table.
// Each user can rate a tour only once (unique constraint on user_id + tour_id).
// Score: 1-5
type Rating struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_user_tour" json:"user_id"`
	TourID    uint      `gorm:"not null;uniqueIndex:idx_user_tour" json:"tour_id"`
	Score     int       `gorm:"not null;check:score >= 1 AND score <= 5" json:"score"`
	Comment   string    `gorm:"type:text" json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Tour Tour `gorm:"foreignKey:TourID" json:"tour,omitempty"`
}

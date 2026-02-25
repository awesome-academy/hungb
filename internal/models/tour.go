package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Tour represents the tours table.
// Status: "draft", "active", or "inactive"
// Images stored as JSON array of URLs.
type Tour struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Title           string         `gorm:"size:500;not null" json:"title"`
	Slug            string         `gorm:"size:500;uniqueIndex;not null" json:"slug"`
	Description     string         `gorm:"type:text" json:"description"`
	Price           float64        `gorm:"type:decimal(15,2);not null" json:"price"`
	DurationDays    int            `gorm:"not null" json:"duration_days"`
	Location        string         `gorm:"size:500" json:"location"`
	MaxParticipants int            `gorm:"not null" json:"max_participants"`
	Images          datatypes.JSON `gorm:"type:json" json:"images"`
	Status          string         `gorm:"size:20;default:'draft';not null" json:"status"`
	AvgRating       float64        `gorm:"type:decimal(3,2);default:0;check:avg_rating >= 0 AND avg_rating <= 5" json:"avg_rating"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Categories []Category     `gorm:"many2many:tour_categories" json:"categories,omitempty"`
	Schedules  []TourSchedule `gorm:"foreignKey:TourID" json:"schedules,omitempty"`
	Bookings   []Booking      `gorm:"foreignKey:TourID" json:"bookings,omitempty"`
	Ratings    []Rating       `gorm:"foreignKey:TourID" json:"ratings,omitempty"`
}

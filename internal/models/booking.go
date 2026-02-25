package models

import (
	"time"
)

// Booking represents the bookings table.
// Status: "pending", "confirmed", "cancelled", or "completed"
type Booking struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"not null;index" json:"user_id"`
	TourID          uint      `gorm:"not null;index" json:"tour_id"`
	ScheduleID      uint      `gorm:"not null;index" json:"schedule_id"`
	NumParticipants int       `gorm:"not null" json:"num_participants"`
	TotalPrice      float64   `gorm:"type:decimal(15,2);not null" json:"total_price"`
	Status          string    `gorm:"size:20;default:'pending';not null" json:"status"`
	Note            string    `gorm:"type:text" json:"note"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Relationships
	User     *User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Tour     *Tour         `gorm:"foreignKey:TourID" json:"tour,omitempty"`
	Schedule *TourSchedule `gorm:"foreignKey:ScheduleID" json:"schedule,omitempty"`
	Payments []Payment     `gorm:"foreignKey:BookingID" json:"payments,omitempty"`
}

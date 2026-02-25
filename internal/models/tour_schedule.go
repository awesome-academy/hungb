package models

import (
	"time"
)

// TourSchedule represents the tour_schedules table.
// Each tour can have multiple departure schedules.
// Status: "open", "full", or "cancelled"
type TourSchedule struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	TourID         uint      `gorm:"not null;index" json:"tour_id"`
	DepartureDate  time.Time `gorm:"not null" json:"departure_date"`
	ReturnDate     time.Time `gorm:"not null" json:"return_date"`
	AvailableSlots int       `gorm:"not null" json:"available_slots"`
	PriceOverride  *float64  `gorm:"type:decimal(15,2)" json:"price_override"`
	Status         string    `gorm:"size:20;default:'open';not null" json:"status"`
	CreatedAt      time.Time `json:"created_at"`

	// Relationships
	Tour     *Tour     `gorm:"foreignKey:TourID" json:"tour,omitempty"`
	Bookings []Booking `gorm:"foreignKey:ScheduleID" json:"bookings,omitempty"`
}

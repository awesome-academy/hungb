package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents the users table.
// Role: "admin" or "user"
// Status: "active", "inactive", or "banned"
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Email     string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"size:255" json:"-"`
	FullName  string         `gorm:"size:255;not null" json:"full_name"`
	Phone     string         `gorm:"size:20" json:"phone"`
	AvatarURL string         `gorm:"size:500" json:"avatar_url"`
	Role      string         `gorm:"size:20;default:user;not null" json:"role"`
	Status    string         `gorm:"size:20;default:active;not null" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	SocialAccounts []SocialAccount `gorm:"foreignKey:UserID" json:"social_accounts,omitempty"`
	BankAccounts   []BankAccount   `gorm:"foreignKey:UserID" json:"bank_accounts,omitempty"`
	Bookings       []Booking       `gorm:"foreignKey:UserID" json:"bookings,omitempty"`
	Ratings        []Rating        `gorm:"foreignKey:UserID" json:"ratings,omitempty"`
	Reviews        []Review        `gorm:"foreignKey:UserID" json:"reviews,omitempty"`
	Comments       []Comment       `gorm:"foreignKey:UserID" json:"comments,omitempty"`
}

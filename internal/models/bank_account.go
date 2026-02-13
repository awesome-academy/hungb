package models

import (
	"time"
)

// BankAccount represents the bank_accounts table.
// Users can have multiple bank accounts, one marked as default for receiving refunds.
type BankAccount struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"not null;index" json:"user_id"`
	BankName      string    `gorm:"size:255;not null" json:"bank_name"`
	AccountNumber string    `gorm:"size:50;not null" json:"account_number"`
	AccountHolder string    `gorm:"size:255;not null" json:"account_holder"`
	IsDefault     bool      `gorm:"default:false" json:"is_default"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

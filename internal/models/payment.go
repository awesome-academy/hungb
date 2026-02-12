package models

import (
	"time"
)

// Payment represents the payments table.
// Status: "pending", "success", "failed", or "refunded"
type Payment struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	BookingID     uint       `gorm:"not null;index" json:"booking_id"`
	BankAccountID *uint      `gorm:"index" json:"bank_account_id"`
	Amount        float64    `gorm:"type:decimal(15,2);not null" json:"amount"`
	PaymentMethod string     `gorm:"size:50" json:"payment_method"`
	TransactionID string     `gorm:"size:255" json:"transaction_id"`
	Status        string     `gorm:"size:20;default:pending;not null" json:"status"`
	PaidAt        *time.Time `json:"paid_at"`
	CreatedAt     time.Time  `json:"created_at"`

	// Relationships
	Booking     Booking      `gorm:"foreignKey:BookingID" json:"booking,omitempty"`
	BankAccount *BankAccount `gorm:"foreignKey:BankAccountID" json:"bank_account,omitempty"`
}

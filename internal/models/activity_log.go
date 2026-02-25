package models

import "time"

type ActivityLog struct {
	ID          uint      `gorm:"primaryKey;type:int unsigned"`
	Action      string    `gorm:"type:varchar(255);not null"`
	UserID      uint      `gorm:"type:int unsigned;not null"`
	Description *string   `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"type:timestamp;autoCreateTime;not null"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

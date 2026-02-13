package models

import "time"

type ActivityLog struct {
	ID          uint      `gorm:"column:id;primaryKey;type:int unsigned"`
	Action      string    `gorm:"column:action;type:varchar(255);not null"`
	UserID      uint      `gorm:"column:user_id;type:int unsigned;not null"`
	Description *string   `gorm:"column:description;type:text"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamp;autoCreateTime;not null"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

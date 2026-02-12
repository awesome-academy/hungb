package models

import (
	"time"
)

// Comment represents the comments table.
// Supports nested replies via ParentID (self-referencing).
type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	ReviewID  uint      `gorm:"not null;index" json:"review_id"`
	ParentID  *uint     `gorm:"index" json:"parent_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	User     User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Review   Review    `gorm:"foreignKey:ReviewID" json:"review,omitempty"`
	Parent   *Comment  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Comment `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

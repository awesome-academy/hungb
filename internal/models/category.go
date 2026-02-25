package models

import (
	"time"

	"gorm.io/gorm"
)

// Category represents the categories table.
// Supports parent-child hierarchy via ParentID (self-referencing).
type Category struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:255;not null" json:"name"`
	Slug        string         `gorm:"size:255;uniqueIndex;not null" json:"slug"`
	Description string         `gorm:"type:text" json:"description"`
	ParentID    *uint          `gorm:"index" json:"parent_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Parent   *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Tours    []Tour     `gorm:"many2many:tour_categories" json:"tours,omitempty"`
}

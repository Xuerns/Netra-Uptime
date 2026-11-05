package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ComponentGroup struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	StatusPageID uuid.UUID `gorm:"type:uuid;not null;index" json:"status_page_id"`
	Name         string    `gorm:"not null" json:"name"`
	DisplayOrder int       `gorm:"default:0" json:"display_order"`
	IsExpanded   bool      `gorm:"default:true" json:"is_expanded"`

	StatusPage StatusPage `gorm:"foreignKey:StatusPageID" json:"status_page,omitempty"`
}

func (m *ComponentGroup) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

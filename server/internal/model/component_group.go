package model

import (
	"github.com/google/uuid"
)

type ComponentGroup struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	StatusPageID uuid.UUID `gorm:"type:uuid;not null;index" json:"status_page_id"`
	Name         string    `gorm:"not null" json:"name"`
	DisplayOrder int       `gorm:"default:0" json:"display_order"`
	IsExpanded   bool      `gorm:"default:true" json:"is_expanded"`

	StatusPage StatusPage `gorm:"foreignKey:StatusPageID" json:"status_page,omitempty"`
}

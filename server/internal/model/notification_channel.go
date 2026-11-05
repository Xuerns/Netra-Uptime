package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationChannel struct {
	ID             uuid.UUID `gorm:"primaryKey;type:uuid"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null"`
	Name           string    `gorm:"not null;default:''"`
	Type           string    `gorm:"not null"` // email | slack | webhook
	Config         string    // stored as JSON string
	IsActive       bool      `gorm:"default:true"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`

	Organization     Organization      `gorm:"foreignKey:OrganizationID"`
	NotificationLogs []NotificationLog `gorm:"foreignKey:ChannelID;constraint:OnDelete:CASCADE;"`
}

func (m *NotificationChannel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

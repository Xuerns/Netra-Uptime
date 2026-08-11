package model

import "github.com/google/uuid"

type NotificationChannel struct {
	ID             uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null"`
	Type           string    `gorm:"not null"` // email | slack | webhook
	Config         string    // disimpan sebagai JSON
	IsActive       bool      `gorm:"default:true"`

	Organization     Organization      `gorm:"foreignKey:OrganizationID"`
	NotificationLogs []NotificationLog `gorm:"foreignKey:ChannelID"`
}

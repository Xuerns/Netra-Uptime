package model

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name      string    `gorm:"not null"`
	Slug      string    `gorm:"unique;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`

	OrganizationMembers  []OrganizationMember  `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	Monitors             []Monitor             `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	StatusPages          []StatusPage          `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	NotificationChannels []NotificationChannel `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	Subscription         *Subscription         `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	AuditLogs            []AuditLog            `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
	APIKeys              []APIKey              `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
}

type OrganizationMember struct {
	ID             uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null"`
	UserID         uuid.UUID `gorm:"type:uuid;not null"`
	Role           string    `gorm:"not null;default:'member'"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`

	Organization Organization `gorm:"foreignKey:OrganizationID"`
	User         User         `gorm:"foreignKey:UserID"`
}

type OrganizationInvitation struct {
	ID             uuid.UUID `gorm:"primartKey;type:uuid;default:gen_random_uuid()"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null"`
	InviterID      uuid.UUID `gorm:"type:uuid;not null"`
	InviteeID      uuid.UUID `gorm:"type:uuid;not null"`
	Role           string    `gorm:"not null;default:'member'"`
	Status         string    `gorm:"not null;default:'pending'"`
	ExpiresAt      time.Time `gorm:"not null"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`

	Organization Organization `gorm:"foreignKey:OrganizationID"`
	Inviter      User         `gorm:"foreignKey:InviterID"`
	Invitee      User         `gorm:"foreignKey:InviteeID"`
}

package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email          string    `gorm:"unique;not null"`
	PasswordHash   *string   // nullable: nil jika login via OAuth
	Name           string    `gorm:"not null"`
	AvatarURL      string    `gorm:"type:text"`
	AuthProvider   string    `gorm:"default:'local'"` // local | google | github
	AuthProviderID string    // ID dari provider (kosong jika local)
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`

	OrganizationMembers []OrganizationMember `gorm:"foreignKey:UserID"`
	RefreshTokens       []RefreshToken       `gorm:"foreignKey:UserID"`
	MonitorPICs         []MonitorPIC         `gorm:"foreignKey:UserID"`
}

package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid"`
	Email        string    `gorm:"unique;not null"`
	PasswordHash *string
	Name         string    `gorm:"not null"`
	AvatarURL    string    `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`

	OrganizationMembers []OrganizationMember `gorm:"foreignKey:UserID"`
	RefreshTokens       []RefreshToken       `gorm:"foreignKey:UserID"`
	MonitorPICs         []MonitorPIC         `gorm:"foreignKey:UserID"`
}

func (m *User) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

package model

import (
	"time"

	"github.com/google/uuid"
)

type PasswordResetToken struct {
	ID        int       `gorm:"primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	Token     string    `gorm:"not null"`
	ExpiredAt time.Time `gorm:"not null"`

	User User `gorm:"foreignKey:UserID;"`
}

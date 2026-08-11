package model

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID              uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrganizationID  uuid.UUID `gorm:"type:uuid;not null"`
	UserID          uuid.UUID `gorm:"type:uuid;not null"`
	OrderID         string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	PlanType        string    `gorm:"type:varchar(50);not null"`
	Amount          int64     `gorm:"not null"`
	Status          string    `gorm:"type:varchar(50);not null;default:'pending'"` // pending | success | failed
	PaymentMethod   string    `gorm:"type:varchar(50)"`
	SnapToken       string    `gorm:"type:varchar(255)"`
	SnapRedirectURL string    `gorm:"type:varchar(255)"`
	CreatedAt       time.Time
	UpdateAt        time.Time

	Organization Organization `gorm:"foreignKey:OrganizationID"`
	User         User         `gorm:"foreignKey:UserID"`
}

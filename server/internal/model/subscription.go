package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	PlanFree = "free"
	PlanPro  = "pro"
	PlanMax  = "max"
)

type Subscription struct {
	ID             uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null;unique"`
	Plan           string     `gorm:"not null;default:'free'"` // free | pro | max
	Status         string     `gorm:"not null;default:'active'"`
	CurrentPeriodEnd *time.Time // nullable

	Organization Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
}

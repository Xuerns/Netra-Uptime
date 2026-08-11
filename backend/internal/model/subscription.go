package model

import (
	"time"

	"github.com/google/uuid"
)

const (
	PlanFree = "free"
	PlanPro  = "pro"
	PlanMax  = "max"
	PricePro = 199000.00
	PriceMax = 299000.00
)

type Subscription struct {
	ID                     uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrganizationID         uuid.UUID `gorm:"type:uuid;not null;unique"`
	Plan                   string    `gorm:"not null;default:'free'"` // hobby | pro | max
	MidtransCustomerID     string
	MidtransSubscriptionID string
	Status                 string     `gorm:"not null;default:'active'"`
	CurrentPeriodEnd       *time.Time // nullable

	Organization Organization `gorm:"foreignKey:OrganizationID;constraint:OnDelete:CASCADE"`
}

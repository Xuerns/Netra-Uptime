package model

import (
	"time"

	"github.com/google/uuid"
)

type Monitor struct {
	ID                 uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OrganizationID     uuid.UUID `gorm:"type:uuid;not null"`
	Name               string    `gorm:"not null"`
	Type               string    `gorm:"not null"` // http | tcp | ping
	Target             string    `gorm:"not null"`
	IntervalSeconds    int       `gorm:"not null;default:60"`
	TimeoutSeconds     int       `gorm:"not null;default:30"`
	ExpectedStatusCode int       `gorm:"default:200"`
	IsActive           bool      `gorm:"default:true"`
	CreatedAt          time.Time `gorm:"autoCreateTime"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`
	Status             string    `gorm:"default:up"`
	FailureCount       int       `gorm:"default:0"`
	Criticality        string    `gorm:"not null;default:'medium'"` // Extrme | High | Medium | Low

	// SSL & Domain Expiry Monitoring
	EnableSSLDomainCheck      bool       `gorm:"default:true"`
	SSLExpiresAt              *time.Time `gorm:"default:null"`
	SSLDaysLeft               *int       `gorm:"default:null"`
	SSLStatus                 string     `gorm:"default:'none'"` // none | valid | warning | expired | error
	SSLError                  string     `gorm:"default:''"`
	DomainExpiresAt           *time.Time `gorm:"default:null"`
	DomainDaysLeft            *int       `gorm:"default:null"`
	DomainStatus              string     `gorm:"default:'none'"` // none | valid | warning | expired | error
	DomainError               string     `gorm:"default:''"`
	LastSSLCheckedAt          *time.Time `gorm:"default:null"`
	LastDomainCheckedAt       *time.Time `gorm:"default:null"`
	LastSSLSentReminderDay    *int       `gorm:"default:null"`
	LastDomainSentReminderDay *int       `gorm:"default:null"`

	Organization       Organization          `gorm:"foreignKey:OrganizationID"`
	Checks             []Check               `gorm:"foreignKey:MonitorID;constraint:OnDelete:CASCADE;"`
	Incidents          []Incident            `gorm:"foreignKey:MonitorID;constraint:OnDelete:CASCADE;"`
	StatusPageMonitors []StatusPageMonitor   `gorm:"foreignKey:MonitorID;constraint:OnDelete:CASCADE;"`
	MonitorPICs        []MonitorPIC          `gorm:"foreignKey:MonitorID;constraint:OnDelete:CASCADE;"`
	NotificationLogs   []NotificationLog     `gorm:"foreignKey:MonitorID;constraint:OnDelete:CASCADE;"`

	RecentChecks        []Check  `gorm:"-" json:"recent_checks"`
	CalculatedUptime30D *float64 `gorm:"-" json:"calculated_uptime_30d"`
}

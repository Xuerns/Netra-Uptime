package model

import (
	"time"

	"github.com/google/uuid"
)

type ScheduledMaintenance struct {
	ID             uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null" json:"organization_id"`
	StatusPageID   uuid.UUID  `gorm:"type:uuid;not null" json:"status_page_id"`
	Title          string     `gorm:"not null" json:"title"`
	Description    string     `gorm:"type:text" json:"description"`
	Status         string     `gorm:"not null;default:'scheduled'" json:"status"` // scheduled | in_progress | completed
	ScheduledStart time.Time  `gorm:"not null" json:"scheduled_start"`
	ScheduledEnd   time.Time  `gorm:"not null" json:"scheduled_end"`
	ActualStart    *time.Time `json:"actual_start,omitempty"`
	ActualEnd      *time.Time `json:"actual_end,omitempty"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	Organization     Organization         `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	StatusPage       StatusPage           `gorm:"foreignKey:StatusPageID" json:"status_page,omitempty"`
	AffectedMonitors []MaintenanceMonitor `gorm:"foreignKey:MaintenanceID;constraint:OnDelete:CASCADE" json:"affected_monitors,omitempty"`
}

type MaintenanceMonitor struct {
	ID            uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	MaintenanceID uuid.UUID `gorm:"type:uuid;not null" json:"maintenance_id"`
	MonitorID     uuid.UUID `gorm:"type:uuid;not null" json:"monitor_id"`

	Maintenance ScheduledMaintenance `gorm:"foreignKey:MaintenanceID" json:"maintenance,omitempty"`
	Monitor     Monitor              `gorm:"foreignKey:MonitorID" json:"monitor,omitempty"`
}

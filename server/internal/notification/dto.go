package notification

import "time"

type CreateChannelRequest struct {
	Name   string            `json:"name" binding:"required,min=2,max=80,nohtml"`
	Type   string            `json:"type" binding:"required,oneof=email slack webhook discord"`
	Config map[string]string `json:"config" binding:"required"`
}

type UpdateChannelRequest struct {
	Name     *string            `json:"name" binding:"omitempty,min=2,max=80,nohtml"`
	IsActive *bool              `json:"is_active"`
	Config   *map[string]string `json:"config" binding:"omitempty"`
}

type ChannelResponse struct {
	ID             string            `json:"id"`
	OrganizationID string            `json:"organization_id"`
	Name           string            `json:"name"`
	Type           string            `json:"type"`   // email | slack | webhook
	Config         map[string]string `json:"config"` // sensitive values are masked
	IsActive       bool              `json:"is_active"`
	CreatedAt      time.Time         `json:"created_at"`
}

type LogResponse struct {
	ID          string     `json:"id"`
	MonitorID   *string    `json:"monitor_id,omitempty"`
	IncidentID  *string    `json:"incident_id,omitempty"`
	ChannelID   *string    `json:"channel_id,omitempty"`
	Channel     string     `json:"channel"`
	Message     string     `json:"message"`
	Status      string     `json:"status"`
	ErrorDetail string     `json:"error_detail,omitempty"`
	SentAt      time.Time  `json:"sent_at"`
}

type LogListResponse struct {
	Logs   []LogResponse `json:"logs"`
	Total  int64         `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

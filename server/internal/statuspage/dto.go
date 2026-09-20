package statuspage

import "time"

type CreateStatusPageRequest struct {
	Subdomain   string   `json:"subdomain" binding:"required,min=3,max=60"`
	Title       string   `json:"title" binding:"required,max=60"`
	Description *string  `json:"description" binding:"omitempty"`
	IsPublic    bool     `json:"is_public" binding:"required"`
	MonitorID []string `json:"monitor_id" binding:"required,min=1,max=40,dive,uuid"`
}

type UpdateStatusPageRequest struct {
	Subdomain   *string   `json:"subdomain" binding:"omitempty,min=3,max=60"`
	Title       *string   `json:"title" binding:"omitempty,max=60"`
	Description *string   `json:"description" binding:"omitempty"`
	IsPublic    *bool     `json:"is_public" binding:"omitempty"`
	MonitorID *[]string `json:"monitor_id" binding:"omitempty,dive,uuid"`
}

type PublicMonitorResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Status       string `json:"status"`
	RecentChecks []int  `json:"recent_checks"`
}

type PublicStatusPageResponse struct {
	Title       string                  `json:"title"`
	Description *string                 `json:"description"`
	Subdomain   string                  `json:"subdomain"`
	Monitors  []PublicMonitorResponse `json:"monitors"`
}

type OrgInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type StatusPageMonitorInfo struct {
	ID           string `json:"id"`
	StatusPageID string `json:"status_page_id"`
	MonitorID    string `json:"monitor_id"`
	DisplayOrder int    `json:"display_order"`
}

type StatusPageResponse struct {
	ID                 string                  `json:"id"`
	OrganizationID     string                  `json:"organization_id"`
	Organization       *OrgInfo                `json:"organization,omitempty"`
	Subdomain          string                  `json:"subdomain"`
	Title              string                  `json:"title"`
	Description        *string                 `json:"description"`
	IsPublic           bool                    `json:"is_public"`
	CreatedAt          time.Time               `json:"created_at"`
	StatusPageMonitors []StatusPageMonitorInfo `json:"status_page_monitors"`
}

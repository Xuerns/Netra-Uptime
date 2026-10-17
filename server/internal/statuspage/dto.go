package statuspage

import "time"

type CreateStatusPageRequest struct {
	Subdomain   string   `json:"subdomain" binding:"required,min=3,max=60"`
	Title       string   `json:"title" binding:"required,max=60"`
	Description *string  `json:"description" binding:"omitempty"`
	IsPublic    bool     `json:"is_public" binding:"required"`
	LogoURL     string   `json:"logo_url" binding:"omitempty"`
	FaviconURL  string   `json:"favicon_url" binding:"omitempty"`
	ThemeColor  string   `json:"theme_color" binding:"omitempty"`
	CustomCSS   string   `json:"custom_css" binding:"omitempty,nohtml,max=10000"`
	MonitorID   []string `json:"monitor_id" binding:"required,min=1,max=40,dive,uuid"`
}

type UpdateStatusPageRequest struct {
	Subdomain   *string   `json:"subdomain" binding:"omitempty,min=3,max=60"`
	Title       *string   `json:"title" binding:"omitempty,max=60"`
	Description *string   `json:"description" binding:"omitempty"`
	IsPublic    *bool     `json:"is_public" binding:"omitempty"`
	LogoURL     *string   `json:"logo_url" binding:"omitempty"`
	FaviconURL  *string   `json:"favicon_url" binding:"omitempty"`
	ThemeColor  *string   `json:"theme_color" binding:"omitempty"`
	CustomCSS   *string   `json:"custom_css" binding:"omitempty,nohtml,max=10000"`
	MonitorID   *[]string `json:"monitor_id" binding:"omitempty,dive,uuid"`
}

type PublicIncidentUpdateResponse struct {
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

type PublicIncidentResponse struct {
	ID          string                         `json:"id"`
	Status      string                         `json:"status"`
	Severity    string                         `json:"severity"`
	Summary     string                         `json:"summary"`
	StartedAt   string                         `json:"started_at"`
	ResolvedAt  *string                        `json:"resolved_at"`
	MonitorName string                         `json:"monitor_name"`
	Updates     []PublicIncidentUpdateResponse `json:"updates"`
}

type PublicUptimeDayResponse struct {
	Date             string  `json:"date"`
	UptimePercentage float64 `json:"uptime_percentage"`
	DownChecks       int     `json:"down_checks"`
	TotalChecks      int     `json:"total_checks"`
	HasData          bool    `json:"has_data"`
}

type PublicMaintenanceResponse struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Status           string   `json:"status"`
	ScheduledStart   string   `json:"scheduled_start"`
	ScheduledEnd     string   `json:"scheduled_end"`
	AffectedMonitors []string `json:"affected_monitors"`
}

type PublicComponentGroupResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	DisplayOrder int    `json:"display_order"`
	IsExpanded   bool   `json:"is_expanded"`
}

type PublicIncidentDayGroup struct {
	Date      string                   `json:"date"`
	Incidents []PublicIncidentResponse `json:"incidents"`
}

type PublicMonitorResponse struct {
	ID            string                    `json:"id"`
	Name          string                    `json:"name"`
	Type          string                    `json:"type"`
	Status        string                    `json:"status"`
	GroupID       *string                   `json:"group_id,omitempty"`
	GroupName     *string                   `json:"group_name,omitempty"`
	UptimePercent *float64                  `json:"uptime_percent"`
	RecentChecks  []int                     `json:"recent_checks"`
	DailyStats    []PublicUptimeDayResponse `json:"daily_stats"`
}

type PublicStatusPageResponse struct {
	Title                string                         `json:"title"`
	Description          *string                        `json:"description"`
	Subdomain            string                         `json:"subdomain"`
	OverallStatus        string                         `json:"overall_status"`
	LogoURL              string                         `json:"logo_url"`
	FaviconURL           string                         `json:"favicon_url"`
	ThemeColor           string                         `json:"theme_color"`
	CustomCSS            string                         `json:"custom_css"`
	Monitors             []PublicMonitorResponse        `json:"monitors"`
	ComponentGroups      []PublicComponentGroupResponse `json:"component_groups"`
	ActiveIncidents      []PublicIncidentResponse       `json:"active_incidents"`
	RecentIncidents      []PublicIncidentResponse       `json:"recent_incidents"`
	IncidentHistory      []PublicIncidentDayGroup       `json:"incident_history"`
	UpcomingMaintenances []PublicMaintenanceResponse    `json:"upcoming_maintenances"`
	OngoingMaintenances  []PublicMaintenanceResponse    `json:"ongoing_maintenances"`
}

type PublicStatusJSONResponse struct {
	Status      string `json:"status"`
	Description string `json:"description"`
	UpdatedAt   string `json:"updated_at"`
}

type CreateMaintenanceRequest struct {
	Title          string    `json:"title" binding:"required,min=3,max=120"`
	Description    string    `json:"description" binding:"omitempty"`
	ScheduledStart time.Time `json:"scheduled_start" binding:"required"`
	ScheduledEnd   time.Time `json:"scheduled_end" binding:"required"`
	MonitorIDs     []string  `json:"monitor_ids" binding:"omitempty,dive,uuid"`
}

type UpdateMaintenanceRequest struct {
	Title          *string    `json:"title" binding:"omitempty,min=3,max=120"`
	Description    *string    `json:"description" binding:"omitempty"`
	Status         *string    `json:"status" binding:"omitempty,oneof=scheduled in_progress completed"`
	ScheduledStart *time.Time `json:"scheduled_start" binding:"omitempty"`
	ScheduledEnd   *time.Time `json:"scheduled_end" binding:"omitempty"`
	MonitorIDs     *[]string  `json:"monitor_ids" binding:"omitempty,dive,uuid"`
}

type MaintenanceResponse struct {
	ID                 string     `json:"id"`
	StatusPageID       string     `json:"status_page_id"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	Status             string     `json:"status"`
	ScheduledStart     time.Time  `json:"scheduled_start"`
	ScheduledEnd       time.Time  `json:"scheduled_end"`
	ActualStart        *time.Time `json:"actual_start,omitempty"`
	ActualEnd          *time.Time `json:"actual_end,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	AffectedMonitors   []string   `json:"affected_monitors"`
	AffectedMonitorIDs []string   `json:"affected_monitor_ids"`
}

type OrgInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type StatusPageMonitorInfo struct {
	ID           string  `json:"id"`
	StatusPageID string  `json:"status_page_id"`
	MonitorID    string  `json:"monitor_id"`
	DisplayOrder int     `json:"display_order"`
	GroupID      *string `json:"group_id,omitempty"`
}

type ComponentGroupResponse struct {
	ID           string   `json:"id"`
	StatusPageID string   `json:"status_page_id"`
	Name         string   `json:"name"`
	DisplayOrder int      `json:"display_order"`
	IsExpanded   bool     `json:"is_expanded"`
	MonitorIDs   []string `json:"monitor_ids"`
}

type CreateComponentGroupRequest struct {
	Name         string `json:"name" binding:"required,min=1,max=100"`
	DisplayOrder int    `json:"display_order"`
	IsExpanded   *bool  `json:"is_expanded"`
}

type UpdateComponentGroupRequest struct {
	Name         *string `json:"name" binding:"omitempty,min=1,max=100"`
	DisplayOrder *int    `json:"display_order"`
	IsExpanded   *bool   `json:"is_expanded"`
}

type AssignMonitorGroupRequest struct {
	GroupID *string `json:"group_id" binding:"omitempty"`
}

type StatusPageResponse struct {
	ID                 string                   `json:"id"`
	OrganizationID     string                   `json:"organization_id"`
	Organization       *OrgInfo                 `json:"organization,omitempty"`
	Subdomain          string                   `json:"subdomain"`
	Title              string                   `json:"title"`
	Description        *string                  `json:"description"`
	IsPublic           bool                     `json:"is_public"`
	LogoURL            string                   `json:"logo_url"`
	FaviconURL         string                   `json:"favicon_url"`
	ThemeColor         string                   `json:"theme_color"`
	CustomCSS          string                   `json:"custom_css"`
	CreatedAt          time.Time                `json:"created_at"`
	StatusPageMonitors []StatusPageMonitorInfo  `json:"status_page_monitors"`
	ComponentGroups    []ComponentGroupResponse `json:"component_groups"`
}

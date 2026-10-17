package monitor

type CreateMonitorRequest struct {
	Name                 string   `json:"name" binding:"required,max=100,min=4,nohtml"`
	Type                 string   `json:"type" binding:"required,oneof=http tcp ping"`
	Target               string   `json:"target" binding:"required,max=100,min=4,nohtml"`
	IntervalSeconds      int      `json:"interval_seconds" binding:"required,min=10"`
	TimeoutSeconds       int      `json:"timeout_seconds" binding:"required,min=1"`
	ExpectedStatusCode   int      `json:"expectation_status_code" binding:"omitempty"`
	Criticality          string   `json:"criticality" binding:"required,oneof=extreme high medium low"`
	MonitorPICs          []string `json:"MonitorPICs" binding:"omitempty,max=5,dive,uuid"`
	Assignees            []string `json:"assignees" binding:"omitempty,max=5,dive,uuid"`
	EnableSSLDomainCheck *bool    `json:"enable_ssl_domain_check" binding:"omitempty"`
}

type UpdatedMonitorRequest struct {
	Name                 *string   `json:"name" binding:"omitempty,max=100,min=4,nohtml"`
	Target               *string   `json:"target" binding:"omitempty,max=100,min=4,nohtml"`
	IntervalSeconds      *int      `json:"interval_seconds" binding:"omitempty,min=10"`
	TimeoutSeconds       *int      `json:"timeout_seconds" binding:"omitempty,min=1"`
	ExpectedStatusCode   *int      `json:"expectation_status_code" binding:"omitempty"`
	IsActive             *bool     `json:"is_active"`
	Criticality          *string   `json:"criticality" binding:"omitempty,oneof=extreme high medium low"`
	MonitorPICs          *[]string `json:"MonitorPICs" binding:"omitempty,dive,uuid"`
	Assignees            *[]string `json:"assignees" binding:"omitempty,dive,uuid"`
	EnableSSLDomainCheck *bool     `json:"enable_ssl_domain_check" binding:"omitempty"`
}

type MonitorResponse struct {
	ID                   string                `json:"id"`
	Name                 string                `json:"name"`
	Type                 string                `json:"type"`
	Target               string                `json:"target"`
	Status               string                `json:"status"`
	IntervalSeconds      int                   `json:"interval_seconds"`
	TimeoutSeconds       int                   `json:"timeout_seconds"`
	ExpectedStatusCode   int                   `json:"expectation_status_code"` // Kept to match binding
	Criticality          string                `json:"criticality"`
	IsActive             bool                  `json:"is_active"`
	EnableSSLDomainCheck bool                  `json:"enable_ssl_domain_check"`
	SSLExpiresAt         *string               `json:"ssl_expires_at,omitempty"`
	SSLDaysLeft          *int                  `json:"ssl_days_left,omitempty"`
	SSLStatus            string                `json:"ssl_status,omitempty"`
	SSLError             string                `json:"ssl_error,omitempty"`
	DomainExpiresAt      *string               `json:"domain_expires_at,omitempty"`
	DomainDaysLeft       *int                  `json:"domain_days_left,omitempty"`
	DomainStatus         string                `json:"domain_status,omitempty"`
	DomainError          string                `json:"domain_error,omitempty"`
	UptimePercent        *float64              `json:"uptime_percent,omitempty"`
	ResponseTimeMs       *int                  `json:"response_time_ms,omitempty"`
	LastCheckedAt        *string               `json:"last_checked_at,omitempty"`
	OrganizationID       string                `json:"organization_id"`
	Organization         *OrgInfo              `json:"organization,omitempty"`
	MonitorPICs          []MonitorPICInfo      `json:"monitor_pics,omitempty"`
	Assignees            []MonitorAssigneeInfo `json:"assignees,omitempty"`
	RecentChecks         []RecentCheck         `json:"recent_checks,omitempty"`
}

type OrgInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type MonitorPICInfo struct {
	UserID string `json:"user_id"`
}

type MonitorAssigneeInfo = MonitorPICInfo

type RecentCheck struct {
	ID             string  `json:"id"`
	MonitorID      string  `json:"monitor_id"`
	Status         string  `json:"status"`
	StatusCode     int     `json:"status_code"`
	ResponseTimeMs int     `json:"response_time_ms"`
	ErrorMessage   *string `json:"error_message,omitempty"`
	CheckedAt      string  `json:"checked_at"`
}

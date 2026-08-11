package audit

type AuditLogResponse struct {
	ID             string  `json:"id"`
	OrganizationID string  `json:"organization_id"`
	UserID         string  `json:"user_id"`
	UserName       string  `json:"user_name,omitempty"`
	UserEmail      string  `json:"user_email,omitempty"`
	Action         string  `json:"action"`
	EntityType     string  `json:"entity_type"`
	EntityID       *string `json:"entity_id,omitempty"`
	Description    string  `json:"description"`
	Details        string  `json:"details,omitempty"`
	IPAddress      string  `json:"ip_address,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

type GetAuditLogsQuery struct {
	Page       int    `form:"page,default=1"`
	Limit      int    `form:"limit,default=20"`
	Action     string `form:"action"`
	EntityType string `form:"entity_type"`
	Search     string `form:"search"`
}

type PaginatedAuditLogsResponse struct {
	Logs       []AuditLogResponse `json:"logs"`
	TotalCount int64              `json:"total_count"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
}

type RecordAuditLogParams struct {
	OrganizationID string
	UserID         string
	Action         string
	EntityType     string
	EntityID       *string
	Description    string
	Details        string
	IPAddress      string
}

package incident

import "time"

type UpdateIncidentRequest struct {
	Message string `json:"message" binding:"required,min=10,max=200"`
}

type GetIncidentUpdateResponse struct {
	Message   string    `json:"message"`
	Username  string    `json:"username"`
	AvatarUrl string    `json:"avatarUrl"`
	CreateAt  time.Time `json:"createAt"`
}

type IncidentResponse struct {
	ID              string             `json:"id"`
	MonitorID       string             `json:"monitor_id"`
	Status          string             `json:"status"`
	Severity        string             `json:"severity"`
	Summary         string             `json:"summary"`
	TriggeredRule   string             `json:"triggered_rule,omitempty"`
	ResponseSnippet string             `json:"response_snippet,omitempty"`
	StartedAt       string             `json:"started_at"`
	ResolvedAt      *string            `json:"resolved_at"`
	Monitor         *IncidentMonitor   `json:"monitor,omitempty"`
	Assignees       []IncidentAssignee `json:"assignees,omitempty"`
}

type IncidentMonitor struct {
	Name   string `json:"name"`
	Target string `json:"target"`
}

type IncidentAssignee struct {
	UserID    string               `json:"user_id"`
	Role      string               `json:"role"`
	CreatedAt string               `json:"created_at"`
	User      IncidentAssigneeUser `json:"user"`
}

type IncidentAssigneeUser struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

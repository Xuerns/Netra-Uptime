package alerttemplate

import (
	"netra-uptime/internal/alertrule"
	"time"

	"github.com/google/uuid"
)

type CreateAlertTemplateRequest struct {
	Name        string                             `json:"name" binding:"required,min=2,max=100,nohtml"`
	Description string                             `json:"description" binding:"omitempty,max=255,nohtml"`
	IsDefault   bool                               `json:"is_default"`
	Rules       []alertrule.CreateAlertRuleRequest `json:"rules" binding:"dive"`
}

type UpdateAlertTemplateRequest struct {
	Name        *string                            `json:"name" binding:"omitempty,min=2,max=100,nohtml"`
	Description *string                            `json:"description" binding:"omitempty,max=255,nohtml"`
	IsDefault   *bool                              `json:"is_default"`
	Rules       []alertrule.CreateAlertRuleRequest `json:"rules" binding:"dive"`
}

type ApplyTemplateRequest struct {
	TemplateID *uuid.UUID `json:"template_id"`
}

type AlertTemplateResponse struct {
	ID             uuid.UUID                     `json:"id"`
	OrganizationID uuid.UUID                     `json:"organization_id"`
	Name           string                        `json:"name"`
	Description    string                        `json:"description"`
	IsDefault      bool                          `json:"is_default"`
	Rules          []alertrule.AlertRuleResponse `json:"rules"`
	MonitorsCount  int64                         `json:"monitors_count"`
	CreatedAt      time.Time                     `json:"created_at"`
	UpdatedAt      time.Time                     `json:"updated_at"`
}

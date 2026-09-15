package alertrule

import (
	"time"

	"github.com/google/uuid"
)

type CreateAlertRuleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100,nohtml"`
	Metric      string `json:"metric" binding:"required,oneof=status_code response_time_ms timeout"`
	Operator    string `json:"operator" binding:"required,oneof=== != > >= < <= in not_in"`
	TargetValue string `json:"target_value" binding:"required,min=1,max=100"`
	Consecutive int    `json:"consecutive" binding:"required,min=1,max=20"`
	Severity    string `json:"severity" binding:"required,oneof=critical warning"`
	IsActive    *bool  `json:"is_active"`
}

type UpdateAlertRuleRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=2,max=100,nohtml"`
	Metric      *string `json:"metric" binding:"omitempty,oneof=status_code response_time_ms timeout"`
	Operator    *string `json:"operator" binding:"omitempty,oneof=== != > >= < <= in not_in"`
	TargetValue *string `json:"target_value" binding:"omitempty,min=1,max=100"`
	Consecutive *int    `json:"consecutive" binding:"omitempty,min=1,max=20"`
	Severity    *string `json:"severity" binding:"omitempty,oneof=critical warning"`
	IsActive    *bool   `json:"is_active"`
}

type AlertRuleResponse struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	MonitorID      *uuid.UUID `json:"monitor_id,omitempty"`
	Name           string     `json:"name"`
	Metric         string     `json:"metric"`
	Operator       string     `json:"operator"`
	TargetValue    string     `json:"target_value"`
	Consecutive    int        `json:"consecutive"`
	Severity       string     `json:"severity"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

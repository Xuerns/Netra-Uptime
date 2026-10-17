package alerttemplate

import (
	"errors"
	"netra-uptime/internal/alertrule"
	"netra-uptime/internal/model"

	"github.com/google/uuid"
)

type Service interface {
	Create(orgID uuid.UUID, req *CreateAlertTemplateRequest) (*AlertTemplateResponse, error)
	GetAll(orgID uuid.UUID) ([]AlertTemplateResponse, error)
	GetByID(orgID, templateID uuid.UUID) (*AlertTemplateResponse, error)
	Update(orgID, templateID uuid.UUID, req *UpdateAlertTemplateRequest) (*AlertTemplateResponse, error)
	Delete(orgID, templateID uuid.UUID) error

	CreateRule(orgID, templateID uuid.UUID, req *alertrule.CreateAlertRuleRequest) (*alertrule.AlertRuleResponse, error)
	UpdateRule(orgID, templateID, ruleID uuid.UUID, req *alertrule.UpdateAlertRuleRequest) (*alertrule.AlertRuleResponse, error)
	DeleteRule(orgID, templateID, ruleID uuid.UUID) error

	AssignToMonitor(orgID, monitorID uuid.UUID, templateID *uuid.UUID) error
}

type serviceImpl struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &serviceImpl{repo: repo}
}

func (s *serviceImpl) Create(orgID uuid.UUID, req *CreateAlertTemplateRequest) (*AlertTemplateResponse, error) {
	tmpl := model.AlertRuleTemplate{
		OrganizationID: orgID,
		Name:           req.Name,
		Description:    req.Description,
		IsDefault:      req.IsDefault,
	}

	for _, r := range req.Rules {
		isActive := true
		if r.IsActive != nil {
			isActive = *r.IsActive
		}
		tmpl.Rules = append(tmpl.Rules, model.AlertRule{
			OrganizationID: orgID,
			Name:           r.Name,
			Metric:         r.Metric,
			Operator:       r.Operator,
			TargetValue:    r.TargetValue,
			Consecutive:    r.Consecutive,
			Severity:       r.Severity,
			IsActive:       isActive,
		})
	}

	if err := s.repo.Create(&tmpl); err != nil {
		return nil, err
	}

	res := mapTemplateToResponse(&tmpl)
	return &res, nil
}

func (s *serviceImpl) GetAll(orgID uuid.UUID) ([]AlertTemplateResponse, error) {
	templates, err := s.repo.GetAll(orgID)
	if err != nil {
		return nil, err
	}

	if len(templates) == 0 {
		defaultTmpl := model.AlertRuleTemplate{
			OrganizationID: orgID,
			Name:           "Standard Web Service",
			Description:    "Default alert policy for web services: 5xx errors and high latency warnings.",
			IsDefault:      true,
			Rules: []model.AlertRule{
				{
					OrganizationID: orgID,
					Name:           "5xx Server Error Alert",
					Metric:         "status_code",
					Operator:       "==",
					TargetValue:    "5xx",
					Consecutive:    2,
					Severity:       "critical",
					IsActive:       true,
				},
				{
					OrganizationID: orgID,
					Name:           "High Latency Warning",
					Metric:         "response_time_ms",
					Operator:       ">",
					TargetValue:    "3000",
					Consecutive:    3,
					Severity:       "warning",
					IsActive:       true,
				},
			},
		}
		if seedErr := s.repo.Create(&defaultTmpl); seedErr == nil {
			templates = append(templates, defaultTmpl)
		}
	}

	var res []AlertTemplateResponse
	for _, t := range templates {
		res = append(res, mapTemplateToResponse(&t))
	}
	return res, nil
}

func (s *serviceImpl) GetByID(orgID, templateID uuid.UUID) (*AlertTemplateResponse, error) {
	tmpl, err := s.repo.GetByID(orgID, templateID)
	if err != nil {
		return nil, errors.New("alert template not found")
	}
	res := mapTemplateToResponse(tmpl)
	return &res, nil
}

func (s *serviceImpl) Update(orgID, templateID uuid.UUID, req *UpdateAlertTemplateRequest) (*AlertTemplateResponse, error) {
	tmpl, err := s.repo.GetByID(orgID, templateID)
	if err != nil {
		return nil, errors.New("alert template not found")
	}

	if req.Name != nil {
		tmpl.Name = *req.Name
	}
	if req.Description != nil {
		tmpl.Description = *req.Description
	}
	if req.IsDefault != nil {
		tmpl.IsDefault = *req.IsDefault
	}

	if err := s.repo.Update(tmpl); err != nil {
		return nil, err
	}

	res := mapTemplateToResponse(tmpl)
	return &res, nil
}

func (s *serviceImpl) Delete(orgID, templateID uuid.UUID) error {
	_, err := s.repo.GetByID(orgID, templateID)
	if err != nil {
		return errors.New("alert template not found")
	}
	return s.repo.Delete(orgID, templateID)
}

func (s *serviceImpl) CreateRule(orgID, templateID uuid.UUID, req *alertrule.CreateAlertRuleRequest) (*alertrule.AlertRuleResponse, error) {
	// Verify template belongs to org
	_, err := s.repo.GetByID(orgID, templateID)
	if err != nil {
		return nil, errors.New("alert template not found")
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	rule := model.AlertRule{
		OrganizationID: orgID,
		TemplateID:     &templateID,
		Name:           req.Name,
		Metric:         req.Metric,
		Operator:       req.Operator,
		TargetValue:    req.TargetValue,
		Consecutive:    req.Consecutive,
		Severity:       req.Severity,
		IsActive:       isActive,
	}

	if err := s.repo.CreateRule(&rule); err != nil {
		return nil, err
	}

	res := mapRuleToResponse(&rule)
	return &res, nil
}

func (s *serviceImpl) UpdateRule(orgID, templateID, ruleID uuid.UUID, req *alertrule.UpdateAlertRuleRequest) (*alertrule.AlertRuleResponse, error) {
	_, err := s.repo.GetByID(orgID, templateID)
	if err != nil {
		return nil, errors.New("alert template not found")
	}

	rule, err := s.repo.GetRuleByID(templateID, ruleID)
	if err != nil {
		return nil, errors.New("rule not found in template")
	}

	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.Metric != nil {
		rule.Metric = *req.Metric
	}
	if req.Operator != nil {
		rule.Operator = *req.Operator
	}
	if req.TargetValue != nil {
		rule.TargetValue = *req.TargetValue
	}
	if req.Consecutive != nil {
		rule.Consecutive = *req.Consecutive
	}
	if req.Severity != nil {
		rule.Severity = *req.Severity
	}
	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateRule(rule); err != nil {
		return nil, err
	}

	res := mapRuleToResponse(rule)
	return &res, nil
}

func (s *serviceImpl) DeleteRule(orgID, templateID, ruleID uuid.UUID) error {
	_, err := s.repo.GetByID(orgID, templateID)
	if err != nil {
		return errors.New("alert template not found")
	}
	return s.repo.DeleteRule(templateID, ruleID)
}

func (s *serviceImpl) AssignToMonitor(orgID, monitorID uuid.UUID, templateID *uuid.UUID) error {
	if templateID != nil {
		// verify template belongs to organization
		if _, err := s.repo.GetByID(orgID, *templateID); err != nil {
			return errors.New("alert template not found or does not belong to organization")
		}
	}
	return s.repo.AssignTemplateToMonitor(orgID, monitorID, templateID)
}

func mapTemplateToResponse(t *model.AlertRuleTemplate) AlertTemplateResponse {
	var ruleResponses []alertrule.AlertRuleResponse
	for _, r := range t.Rules {
		ruleResponses = append(ruleResponses, mapRuleToResponse(&r))
	}

	return AlertTemplateResponse{
		ID:             t.ID,
		OrganizationID: t.OrganizationID,
		Name:           t.Name,
		Description:    t.Description,
		IsDefault:      t.IsDefault,
		Rules:          ruleResponses,
		MonitorsCount:  t.MonitorsCount,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}

func mapRuleToResponse(r *model.AlertRule) alertrule.AlertRuleResponse {
	return alertrule.AlertRuleResponse{
		ID:             r.ID,
		OrganizationID: r.OrganizationID,
		MonitorID:      r.MonitorID,
		Name:           r.Name,
		Metric:         r.Metric,
		Operator:       r.Operator,
		TargetValue:    r.TargetValue,
		Consecutive:    r.Consecutive,
		Severity:       r.Severity,
		IsActive:       r.IsActive,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}

package alertrule

import (
	"PulseWatch/internal/model"
	"errors"

	"github.com/google/uuid"
)

type Service interface {
	Create(orgID, monitorID uuid.UUID, req *CreateAlertRuleRequest) (*AlertRuleResponse, error)
	GetByMonitor(orgID, monitorID uuid.UUID) ([]AlertRuleResponse, error)
	GetByID(orgID, ruleID uuid.UUID) (*AlertRuleResponse, error)
	Update(orgID, ruleID uuid.UUID, req *UpdateAlertRuleRequest) (*AlertRuleResponse, error)
	Delete(orgID, ruleID uuid.UUID) error
}

type serviceImpl struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &serviceImpl{repo: repo}
}

func (s *serviceImpl) Create(orgID, monitorID uuid.UUID, req *CreateAlertRuleRequest) (*AlertRuleResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	rule := model.AlertRule{
		OrganizationID: orgID,
		MonitorID:      &monitorID,
		Name:           req.Name,
		Metric:         req.Metric,
		Operator:       req.Operator,
		TargetValue:    req.TargetValue,
		Consecutive:    req.Consecutive,
		Severity:       req.Severity,
		IsActive:       isActive,
	}

	if err := s.repo.Create(&rule); err != nil {
		return nil, err
	}

	res := mapToResponse(&rule)
	return &res, nil
}

func (s *serviceImpl) GetByMonitor(orgID, monitorID uuid.UUID) ([]AlertRuleResponse, error) {
	rules, err := s.repo.GetByMonitor(orgID, monitorID)
	if err != nil {
		return nil, err
	}

	var responses []AlertRuleResponse
	for _, r := range rules {
		responses = append(responses, mapToResponse(&r))
	}
	return responses, nil
}

func (s *serviceImpl) GetByID(orgID, ruleID uuid.UUID) (*AlertRuleResponse, error) {
	rule, err := s.repo.GetByID(orgID, ruleID)
	if err != nil {
		return nil, errors.New("Alert rule not found")
	}
	res := mapToResponse(rule)
	return &res, nil
}

func (s *serviceImpl) Update(orgID, ruleID uuid.UUID, req *UpdateAlertRuleRequest) (*AlertRuleResponse, error) {
	rule, err := s.repo.GetByID(orgID, ruleID)
	if err != nil {
		return nil, errors.New("Alert rule not found")
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

	if err := s.repo.Update(rule); err != nil {
		return nil, err
	}

	res := mapToResponse(rule)
	return &res, nil
}

func (s *serviceImpl) Delete(orgID, ruleID uuid.UUID) error {
	return s.repo.Delete(orgID, ruleID)
}

func mapToResponse(r *model.AlertRule) AlertRuleResponse {
	return AlertRuleResponse{
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

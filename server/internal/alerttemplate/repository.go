package alerttemplate

import (
	"netra-uptime/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(template *model.AlertRuleTemplate) error
	GetAll(orgID uuid.UUID) ([]model.AlertRuleTemplate, error)
	GetByID(orgID, templateID uuid.UUID) (*model.AlertRuleTemplate, error)
	Update(template *model.AlertRuleTemplate) error
	Delete(orgID, templateID uuid.UUID) error

	CreateRule(rule *model.AlertRule) error
	GetRules(templateID uuid.UUID) ([]model.AlertRule, error)
	GetRuleByID(templateID, ruleID uuid.UUID) (*model.AlertRule, error)
	UpdateRule(rule *model.AlertRule) error
	DeleteRule(templateID, ruleID uuid.UUID) error

	AssignTemplateToMonitor(orgID, monitorID uuid.UUID, templateID *uuid.UUID) error
}

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) Create(template *model.AlertRuleTemplate) error {
	return r.db.Create(template).Error
}

func (r *repositoryImpl) GetAll(orgID uuid.UUID) ([]model.AlertRuleTemplate, error) {
	var templates []model.AlertRuleTemplate
	err := r.db.Preload("Rules", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Where("organization_id = ?", orgID).
		Order("created_at DESC").
		Find(&templates).Error
	if err != nil {
		return nil, err
	}

	for i := range templates {
		var count int64
		r.db.Model(&model.Monitor{}).Where("alert_template_id = ?", templates[i].ID).Count(&count)
		templates[i].MonitorsCount = count
	}

	return templates, nil
}

func (r *repositoryImpl) GetByID(orgID, templateID uuid.UUID) (*model.AlertRuleTemplate, error) {
	var template model.AlertRuleTemplate
	err := r.db.Preload("Rules", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Where("organization_id = ? AND id = ?", orgID, templateID).First(&template).Error
	if err != nil {
		return nil, err
	}

	var count int64
	r.db.Model(&model.Monitor{}).Where("alert_template_id = ?", template.ID).Count(&count)
	template.MonitorsCount = count

	return &template, nil
}

func (r *repositoryImpl) Update(template *model.AlertRuleTemplate) error {
	return r.db.Save(template).Error
}

func (r *repositoryImpl) Delete(orgID, templateID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Nullify template link on any monitors
		if err := tx.Model(&model.Monitor{}).Where("alert_template_id = ?", templateID).Update("alert_template_id", nil).Error; err != nil {
			return err
		}
		// Delete template rules
		if err := tx.Where("template_id = ?", templateID).Delete(&model.AlertRule{}).Error; err != nil {
			return err
		}
		// Delete template
		return tx.Where("organization_id = ? AND id = ?", orgID, templateID).Delete(&model.AlertRuleTemplate{}).Error
	})
}

func (r *repositoryImpl) CreateRule(rule *model.AlertRule) error {
	return r.db.Create(rule).Error
}

func (r *repositoryImpl) GetRules(templateID uuid.UUID) ([]model.AlertRule, error) {
	var rules []model.AlertRule
	err := r.db.Where("template_id = ?", templateID).Order("created_at ASC").Find(&rules).Error
	return rules, err
}

func (r *repositoryImpl) GetRuleByID(templateID, ruleID uuid.UUID) (*model.AlertRule, error) {
	var rule model.AlertRule
	err := r.db.Where("template_id = ? AND id = ?", templateID, ruleID).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *repositoryImpl) UpdateRule(rule *model.AlertRule) error {
	return r.db.Save(rule).Error
}

func (r *repositoryImpl) DeleteRule(templateID, ruleID uuid.UUID) error {
	return r.db.Where("template_id = ? AND id = ?", templateID, ruleID).Delete(&model.AlertRule{}).Error
}

func (r *repositoryImpl) AssignTemplateToMonitor(orgID, monitorID uuid.UUID, templateID *uuid.UUID) error {
	return r.db.Model(&model.Monitor{}).
		Where("organization_id = ? AND id = ?", orgID, monitorID).
		Update("alert_template_id", templateID).Error
}

package alertrule

import (
	"PulseWatch/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(rule *model.AlertRule) error
	GetByMonitor(orgID, monitorID uuid.UUID) ([]model.AlertRule, error)
	GetByID(orgID, ruleID uuid.UUID) (*model.AlertRule, error)
	Update(rule *model.AlertRule) error
	Delete(orgID, ruleID uuid.UUID) error
}

type repositoryImpl struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) Create(rule *model.AlertRule) error {
	return r.db.Create(rule).Error
}

func (r *repositoryImpl) GetByMonitor(orgID, monitorID uuid.UUID) ([]model.AlertRule, error) {
	var rules []model.AlertRule
	err := r.db.Where("organization_id = ? AND monitor_id = ?", orgID, monitorID).
		Order("created_at ASC").
		Find(&rules).Error
	return rules, err
}

func (r *repositoryImpl) GetByID(orgID, ruleID uuid.UUID) (*model.AlertRule, error) {
	var rule model.AlertRule
	err := r.db.Where("organization_id = ? AND id = ?", orgID, ruleID).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *repositoryImpl) Update(rule *model.AlertRule) error {
	return r.db.Save(rule).Error
}

func (r *repositoryImpl) Delete(orgID, ruleID uuid.UUID) error {
	return r.db.Where("organization_id = ? AND id = ?", orgID, ruleID).Delete(&model.AlertRule{}).Error
}

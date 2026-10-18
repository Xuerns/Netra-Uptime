package incident

import (
	"netra-uptime/internal/model"

	"gorm.io/gorm"
)

type IncidentRepository interface {
	GetById(orgId, incidentId string) (*model.Incident, error)
	GetAllByOrg(orgId string, monitorId ...string) ([]model.Incident, error)
	CreateUpdate(update *model.IncidentUpdate) error
	GetIncidentOnly(incidentId string) (*model.Incident, error)
	AddAssignee(assignee *model.IncidentAssignee) error
	GetAssignee(userId, incidentId string) (*model.IncidentAssignee, error)
	GetMessageUpdate(incidentId string) ([]model.IncidentUpdate, error)
}

type IncidentRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorRepository(db *gorm.DB) IncidentRepository {
	return &IncidentRepositoryImpl{db: db}
}

func (r *IncidentRepositoryImpl) GetById(orgId, incidentId string) (*model.Incident, error) {
	var incident model.Incident
	if err := r.db.Preload("Monitor").Preload("IncidentUpdates.User").Preload("Assignees.User").
		Joins("JOIN monitors ON monitors.id = incidents.monitor_id").
		Where("monitors.organization_id = ? AND incidents.id = ?", orgId, incidentId).First(&incident).Error; err != nil {
		return nil, err
	}
	return &incident, nil
}

func (r *IncidentRepositoryImpl) GetAllByOrg(orgid string, monitorId ...string) ([]model.Incident, error) {
	query := r.db.Preload("Monitor").
		Preload("Assignees.User").
		Preload("IncidentUpdates.User").
		Joins("JOIN monitors ON monitors.id = incidents.monitor_id").
		Where("monitors.organization_id = ?", orgid)

	if len(monitorId) > 0 && monitorId[0] != "" {
		query = query.Where("incidents.monitor_id = ?", monitorId[0])
	}

	var incidents []model.Incident
	if err := query.Order("incidents.started_at DESC").Find(&incidents).Error; err != nil {
		return nil, err
	}
	return incidents, nil
}

func (r *IncidentRepositoryImpl) CreateUpdate(update *model.IncidentUpdate) error {
	return r.db.Create(update).Error
}

func (r *IncidentRepositoryImpl) GetIncidentOnly(incidentId string) (*model.Incident, error) {
	var incident model.Incident
	if err := r.db.Where("id = ?", incidentId).First(&incident).Error; err != nil {
		return nil, err
	}
	return &incident, nil
}

func (r *IncidentRepositoryImpl) AddAssignee(assignee *model.IncidentAssignee) error {
	return r.db.Create(assignee).Error
}

func (r *IncidentRepositoryImpl) GetAssignee(userId, incidentId string) (*model.IncidentAssignee, error) {
	var assignee model.IncidentAssignee
	if err := r.db.Where("user_id = ? AND incident_id = ?", userId, incidentId).First(&assignee).Error; err != nil {
		return nil, err
	}
	return &assignee, nil
}

func (r *IncidentRepositoryImpl) GetMessageUpdate(incidentId string) ([]model.IncidentUpdate, error) {
	var incidentUpdates []model.IncidentUpdate
	if err := r.db.Preload("User").Where("incident_id = ?", incidentId).Order("created_at DESC").Find(&incidentUpdates).Error; err != nil {
		return nil, err
	}
	return incidentUpdates, nil
}

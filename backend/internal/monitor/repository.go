package monitor

import (
	"PulseWatch/config"
	"PulseWatch/internal/model"

	"gorm.io/gorm"
)

type MonitorRepository interface {
	Create(monitor *model.Monitor) error
	GetById(orgId, monitorId string) (*model.Monitor, error)
	GetAllByOrg(orgId string) ([]model.Monitor, error)
	Update(monitor *model.Monitor) error
	Delete(orgId, monitorId string) error
	GetActiveMonitor() ([]model.Monitor, error)
	GetAllMonitorByUserId(userId string) ([]model.Monitor, error)
	ValidateUsersInOrg(orgId string, userIds []string) bool
	CreateMonitorPIC(monitorPIC *model.MonitorPIC) error
	ReplaceMonitorPICs(monitor *model.Monitor, newPICs []model.MonitorPIC) error
	GetNumberOfPIC(orgId, monitorId string) (int, error)
	TotalOrganizationMonitor(orgId string) (int64, error)
}

type MonitorRepositoryImpl struct {
	db *gorm.DB
}

func NewMonitorRepository(db *gorm.DB) MonitorRepository {
	return &MonitorRepositoryImpl{db: db}
}

func (r *MonitorRepositoryImpl) Create(monitor *model.Monitor) error {
	return r.db.Create(&monitor).Error
}

func (r *MonitorRepositoryImpl) GetById(orgId, monitorId string) (*model.Monitor, error) {
	var monitor model.Monitor

	if err := r.db.Where("organization_id = ? AND id = ?", orgId, monitorId).First(&monitor).Error; err != nil {
		return nil, err
	}

	return &monitor, nil
}

func (r *MonitorRepositoryImpl) GetAllByOrg(orgId string) ([]model.Monitor, error) {
	var monitors []model.Monitor

	if err := r.db.Where("organization_id = ?", orgId).Preload("MonitorPICs").Find(&monitors).Error; err != nil {
		return nil, err
	}

	var monitorIds []string
	for _, monitor := range monitors {
		monitorIds = append(monitorIds, monitor.ID.String())
	}

	if len(monitorIds) > 0 {
		limit := config.Load().RecentChecksLimit
		var recentCheck []model.Check
		r.db.Raw(`
			SELECT * FROM (
				SELECT *, ROW_NUMBER() OVER(PARTITION BY monitor_id ORDER BY checked_at DESC) as rn
				FROM checks
				WHERE monitor_id IN ?
			) as subquery
			 WHERE rn <= ? ORDER BY checked_at DESC
		`, monitorIds, limit).Scan(&recentCheck)

		checkMap := make(map[string][]model.Check)
		for _, check := range recentCheck {
			monitorIDStr := check.MonitorID.String()
			checkMap[monitorIDStr] = append(checkMap[monitorIDStr], check)
		}

		for i, monitor := range monitors {
			monitors[i].RecentChecks = checkMap[monitor.ID.String()]
		}
	}

	return monitors, nil
}

func (r *MonitorRepositoryImpl) Update(monitor *model.Monitor) error {
	return r.db.Save(monitor).Error
}

func (r *MonitorRepositoryImpl) Delete(orgId, monitorId string) error {
	return r.db.Where("organization_id = ? AND id = ?", orgId, monitorId).Delete(&model.Monitor{}).Error
}

func (r *MonitorRepositoryImpl) GetActiveMonitor() ([]model.Monitor, error) {
	var monitors []model.Monitor
	if err := r.db.Where("is_active = true").Find(&monitors).Error; err != nil {
		return nil, err
	}
	return monitors, nil
}

func (r *MonitorRepositoryImpl) GetAllMonitorByUserId(userId string) ([]model.Monitor, error) {
	var monitors []model.Monitor
	err := r.db.Preload("Organization").
		Joins("JOIN organization_members ON monitors.organization_id = organization_members.organization_id").
		Where("organization_members.user_id = ?", userId).
		Find(&monitors).Error

	if err != nil {
		return nil, err
	}

	var MonitorIDs []string
	for _, monitor := range monitors {
		MonitorIDs = append(MonitorIDs, monitor.ID.String())
	}

	if len(MonitorIDs) > 0 {
		limit := config.Load().RecentChecksLimit
		var recentChecks []model.Check
		r.db.Raw(`
			Select * FROM (
				SELECT *, ROW_NUMBER() OVER(PARTITION BY monitor_id ORDER BY checked_at DESC) as rn
				FROM checks
				WHERE monitor_id IN ?
			) as subquery
			WHERE rn <= ? ORDER BY checked_at DESC
		`, MonitorIDs, limit).Scan(&recentChecks)

		checkMap := make(map[string][]model.Check)
		for _, check := range recentChecks {
			monitorIDStr := check.MonitorID.String()
			checkMap[monitorIDStr] = append(checkMap[monitorIDStr], check)
		}

		for i, Monitor := range monitors {
			monitors[i].RecentChecks = checkMap[Monitor.ID.String()]
		}
	}

	return monitors, nil
}

func (r *MonitorRepositoryImpl) ValidateUsersInOrg(orgId string, userIds []string) bool {
	var count int64

	r.db.Model(&model.OrganizationMember{}).
		Where("organization_id = ? AND user_id IN ?", orgId, userIds).
		Count(&count)

	return int(count) == len(userIds)
}

func (r *MonitorRepositoryImpl) CreateMonitorPIC(monitorPIC *model.MonitorPIC) error {
	return r.db.Create(&monitorPIC).Error
}

func (r *MonitorRepositoryImpl) ReplaceMonitorPICs(monitor *model.Monitor, newPICs []model.MonitorPIC) error {
	if err := r.db.Where("monitor_id = ?", monitor.ID).Delete(&model.MonitorPIC{}).Error; err != nil {
		return err
	}

	if len(newPICs) > 0 {
		if err := r.db.Model(monitor).Association("MonitorPICs").Append(newPICs); err != nil {
			return err
		}
	}
	return nil
}

func (r *MonitorRepositoryImpl) GetNumberOfPIC(orgId, monitorId string) (int, error) {
	var count int64
	err := r.db.Model(&model.MonitorPIC{}).Where("organization_id = ? AND monitor_id = ?", orgId, monitorId).Count(&count).Error

	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *MonitorRepositoryImpl) TotalOrganizationMonitor(orgId string) (int64, error) {
	var totalOrganizationMonitor int64
	if err := r.db.Model(&model.Monitor{}).Where("organization_id = ?", orgId).Count(&totalOrganizationMonitor).Error; err != nil {
		return 0, err
	}
	return totalOrganizationMonitor, nil
}

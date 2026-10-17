package monitor

import (
	"netra-uptime/config"
	"netra-uptime/internal/model"
	"sort"
	"time"

	"github.com/google/uuid"
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
	GetDailyStats(monitorId string, days int) ([]model.MonitorDailyStat, error)
	Get30DayUptimeMap(monitorIDs []string) (map[string]float64, error)
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

	if err := r.db.Where("organization_id = ? AND id = ?", orgId, monitorId).
		Preload("MonitorPICs").
		Preload("Organization").
		First(&monitor).Error; err != nil {
		return nil, err
	}

	limit := config.Load().RecentChecksLimit
	var recentChecks []model.Check
	if err := r.db.Where("monitor_id = ?", monitor.ID).
		Order("checked_at DESC").
		Limit(limit).
		Find(&recentChecks).Error; err == nil {
		monitor.RecentChecks = recentChecks
	}

	uptimeMap, _ := r.Get30DayUptimeMap([]string{monitor.ID.String()})
	if pct, ok := uptimeMap[monitor.ID.String()]; ok {
		val := pct
		monitor.CalculatedUptime30D = &val
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

		uptimeMap, _ := r.Get30DayUptimeMap(monitorIds)
		for i, monitor := range monitors {
			monitors[i].RecentChecks = checkMap[monitor.ID.String()]
			if pct, ok := uptimeMap[monitor.ID.String()]; ok {
				val := pct
				monitors[i].CalculatedUptime30D = &val
			}
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

		uptimeMap, _ := r.Get30DayUptimeMap(MonitorIDs)
		for i, Monitor := range monitors {
			monitors[i].RecentChecks = checkMap[Monitor.ID.String()]
			if pct, ok := uptimeMap[Monitor.ID.String()]; ok {
				val := pct
				monitors[i].CalculatedUptime30D = &val
			}
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
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("monitor_id = ?", monitor.ID).Delete(&model.MonitorPIC{}).Error; err != nil {
			return err
		}

		if len(newPICs) == 0 {
			return nil
		}

		for i := range newPICs {
			newPICs[i].MonitorID = monitor.ID
		}

		// Single multi-row INSERT instead of one query per PIC.
		return tx.CreateInBatches(&newPICs, 100).Error
	})
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

func (r *MonitorRepositoryImpl) GetDailyStats(monitorId string, days int) ([]model.MonitorDailyStat, error) {
	var stats []model.MonitorDailyStat
	now := time.Now().UTC()
	startDate := time.Date(now.Year(), now.Month(), now.Day()-days, 0, 0, 0, 0, time.UTC)
	err := r.db.Where("monitor_id = ? AND date >= ?", monitorId, startDate).Order("date ASC").Find(&stats).Error
	if err != nil {
		return nil, err
	}

	type RawDailyAgg struct {
		Date         time.Time `gorm:"column:date"`
		TotalChecks  int       `gorm:"column:total_checks"`
		UpChecks     int       `gorm:"column:up_checks"`
		DownChecks   int       `gorm:"column:down_checks"`
		AvgLatencyMs int       `gorm:"column:avg_latency_ms"`
		MinLatencyMs int       `gorm:"column:min_latency_ms"`
		MaxLatencyMs int       `gorm:"column:max_latency_ms"`
		Count2xx     int       `gorm:"column:count_2xx"`
		Count3xx     int       `gorm:"column:count_3xx"`
		Count4xx     int       `gorm:"column:count_4xx"`
		Count5xx     int       `gorm:"column:count_5xx"`
		CountError   int       `gorm:"column:count_error"`
	}

	var rawAggs []RawDailyAgg
	r.db.Table("checks").
		Select(`
			DATE(checked_at) AS date,
			COUNT(*) AS total_checks,
			COUNT(CASE WHEN status = 'up' THEN 1 END) AS up_checks,
			COUNT(CASE WHEN status = 'down' THEN 1 END) AS down_checks,
			ROUND(AVG(response_time_ms)) AS avg_latency_ms,
			MIN(response_time_ms) AS min_latency_ms,
			MAX(response_time_ms) AS max_latency_ms,
			COUNT(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 END) AS count_2xx,
			COUNT(CASE WHEN status_code >= 300 AND status_code < 400 THEN 1 END) AS count_3xx,
			COUNT(CASE WHEN status_code >= 400 AND status_code < 500 THEN 1 END) AS count_4xx,
			COUNT(CASE WHEN status_code >= 500 AND status_code < 600 THEN 1 END) AS count_5xx,
			COUNT(CASE WHEN status_code = 0 OR status_code IS NULL THEN 1 END) AS count_error
		`).
		Where("monitor_id = ? AND checked_at >= ?", monitorId, startDate).
		Group("DATE(checked_at)").
		Scan(&rawAggs)

	if len(rawAggs) > 0 {
		mUUID, _ := uuid.Parse(monitorId)
		dateMap := make(map[string]int)
		for idx, s := range stats {
			dateMap[s.Date.Format("2006-01-02")] = idx
		}

		for _, raw := range rawAggs {
			dateStr := raw.Date.Format("2006-01-02")
			if idx, found := dateMap[dateStr]; found {
				tot := stats[idx].TotalChecks + raw.TotalChecks
				up := stats[idx].UpChecks + raw.UpChecks
				down := stats[idx].DownChecks + raw.DownChecks
				pct := 100.0
				if tot > 0 {
					pct = (float64(up) / float64(tot)) * 100.0
				}
				prevTot := stats[idx].TotalChecks
				stats[idx].TotalChecks = tot
				stats[idx].UpChecks = up
				stats[idx].DownChecks = down
				stats[idx].UptimePercentage = pct
				stats[idx].Count2xx += raw.Count2xx
				stats[idx].Count3xx += raw.Count3xx
				stats[idx].Count4xx += raw.Count4xx
				stats[idx].Count5xx += raw.Count5xx
				stats[idx].CountError += raw.CountError
				if raw.MinLatencyMs > 0 && (stats[idx].MinLatencyMs == 0 || raw.MinLatencyMs < stats[idx].MinLatencyMs) {
					stats[idx].MinLatencyMs = raw.MinLatencyMs
				}
				if raw.MaxLatencyMs > stats[idx].MaxLatencyMs {
					stats[idx].MaxLatencyMs = raw.MaxLatencyMs
				}
				if tot > 0 {
					stats[idx].AvgLatencyMs = (stats[idx].AvgLatencyMs*prevTot + raw.AvgLatencyMs*raw.TotalChecks) / tot
				}
			} else {
				pct := 100.0
				if raw.TotalChecks > 0 {
					pct = (float64(raw.UpChecks) / float64(raw.TotalChecks)) * 100.0
				}
				stats = append(stats, model.MonitorDailyStat{
					MonitorID:        mUUID,
					Date:             raw.Date,
					TotalChecks:      raw.TotalChecks,
					UpChecks:         raw.UpChecks,
					DownChecks:       raw.DownChecks,
					UptimePercentage: pct,
					AvgLatencyMs:     raw.AvgLatencyMs,
					MinLatencyMs:     raw.MinLatencyMs,
					MaxLatencyMs:     raw.MaxLatencyMs,
					Count2xx:         raw.Count2xx,
					Count3xx:         raw.Count3xx,
					Count4xx:         raw.Count4xx,
					Count5xx:         raw.Count5xx,
					CountError:       raw.CountError,
				})
			}
		}

		sort.Slice(stats, func(i, j int) bool {
			return stats[i].Date.Before(stats[j].Date)
		})
	}

	return stats, nil
}

type UptimeAgg struct {
	MonitorID   string `gorm:"column:monitor_id"`
	TotalChecks int64  `gorm:"column:total_checks"`
	UpChecks    int64  `gorm:"column:up_checks"`
}

func (r *MonitorRepositoryImpl) Get30DayUptimeMap(monitorIDs []string) (map[string]float64, error) {
	result := make(map[string]float64)
	if len(monitorIDs) == 0 {
		return result, nil
	}

	sinceDate := time.Now().AddDate(0, 0, -30)

	var dailyAggs []UptimeAgg
	r.db.Table("monitor_daily_stats").
		Select("monitor_id::text as monitor_id, SUM(total_checks) as total_checks, SUM(up_checks) as up_checks").
		Where("monitor_id IN ? AND date >= ?", monitorIDs, sinceDate).
		Group("monitor_id").
		Scan(&dailyAggs)

	var checkAggs []UptimeAgg
	r.db.Table("checks").
		Select("monitor_id::text as monitor_id, COUNT(*) as total_checks, COUNT(CASE WHEN status = 'up' THEN 1 END) as up_checks").
		Where("monitor_id IN ? AND checked_at >= ?", monitorIDs, sinceDate).
		Group("monitor_id").
		Scan(&checkAggs)

	totals := make(map[string]int64)
	ups := make(map[string]int64)

	for _, d := range dailyAggs {
		totals[d.MonitorID] += d.TotalChecks
		ups[d.MonitorID] += d.UpChecks
	}

	for _, c := range checkAggs {
		totals[c.MonitorID] += c.TotalChecks
		ups[c.MonitorID] += c.UpChecks
	}

	for _, id := range monitorIDs {
		tot := totals[id]
		u := ups[id]
		if tot > 0 {
			result[id] = (float64(u) / float64(tot)) * 100.0
		} else {
			result[id] = 100.0
		}
	}

	return result, nil
}

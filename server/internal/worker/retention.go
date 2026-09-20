package worker

import (
	"PulseWatch/internal/model"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RetentionJob struct {
	db       *gorm.DB
	stopChan chan struct{}
}

func NewRetentionJob(db *gorm.DB) *RetentionJob {
	return &RetentionJob{
		db:       db,
		stopChan: make(chan struct{}),
	}
}

func (r RetentionJob) Start() {
	ticker := time.NewTicker(1 * time.Hour)

	go r.CleanOldChecks()

	go func() {
		for {
			select {
			case <-ticker.C:
				r.CleanOldChecks()
			case <-r.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (r *RetentionJob) Stop() {
	close(r.stopChan)
}

func (r *RetentionJob) CleanOldChecks() {
	threshold := time.Now().Add(-24 * time.Hour)

	log.Printf("[Retention] Starting aggregation and cleanup for checks before: %s", threshold.Format("2006-01-02 15:04:05"))

	type AggregateResult struct {
		MonitorID    string    `gorm:"column:monitor_id"`
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

	var results []AggregateResult

	query := `
		SELECT 
			monitor_id,
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
		FROM checks
		WHERE checked_at < ?
		GROUP BY monitor_id, DATE(checked_at)
	`

	if err := r.db.Raw(query, threshold).Scan(&results).Error; err != nil {
		log.Printf("[Retention] Failed to aggregate old checks: %v", err)
	} else if len(results) > 0 {
		log.Printf("[Retention] Aggregating %d daily stat bucket(s)...", len(results))

		for _, agg := range results {
			monitorUUID, err := uuid.Parse(agg.MonitorID)
			if err != nil {
				continue
			}

			uptimePct := 100.0
			if agg.TotalChecks > 0 {
				uptimePct = (float64(agg.UpChecks) / float64(agg.TotalChecks)) * 100.0
			}

			stat := model.MonitorDailyStat{
				MonitorID:        monitorUUID,
				Date:             agg.Date,
				TotalChecks:      agg.TotalChecks,
				UpChecks:         agg.UpChecks,
				DownChecks:       agg.DownChecks,
				UptimePercentage: uptimePct,
				AvgLatencyMs:     agg.AvgLatencyMs,
				MinLatencyMs:     agg.MinLatencyMs,
				MaxLatencyMs:     agg.MaxLatencyMs,
				Count2xx:         agg.Count2xx,
				Count3xx:         agg.Count3xx,
				Count4xx:         agg.Count4xx,
				Count5xx:         agg.Count5xx,
				CountError:       agg.CountError,
			}

			if err := r.db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "monitor_id"}, {Name: "date"}},
				DoUpdates: clause.AssignmentColumns([]string{"total_checks", "up_checks", "down_checks", "uptime_percentage", "avg_latency_ms", "min_latency_ms", "max_latency_ms", "count_2xx", "count_3xx", "count_4xx", "count_5xx", "count_error", "updated_at"}),
			}).Create(&stat).Error; err != nil {
				log.Printf("[Retention] Failed to upsert daily stat for monitor %s on date %s: %v", agg.MonitorID, agg.Date.Format("2006-01-02"), err)
			}
		}
	}

	result := r.db.Where("checked_at < ?", threshold).Delete(&model.Check{})
	if result.Error != nil {
		log.Printf("[Retention] Failed to clean old data check: %v", result.Error)
		return
	}

	log.Printf("[Retention] Success clean old data check: %d row(s)", result.RowsAffected)
}

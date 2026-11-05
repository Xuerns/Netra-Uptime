package worker

import (
	"log"
	"netra-uptime/internal/model"
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
	threshold := time.Now().UTC().Add(-24 * time.Hour)

	log.Printf("[Retention] Starting aggregation and cleanup for checks before: %s", threshold.Format("2006-01-02 15:04:05"))

	type AggregateResult struct {
		MonitorID    string `gorm:"column:monitor_id"`
		Date         string `gorm:"column:date"`
		TotalChecks  int    `gorm:"column:total_checks"`
		UpChecks     int    `gorm:"column:up_checks"`
		DownChecks   int    `gorm:"column:down_checks"`
		AvgLatencyMs int    `gorm:"column:avg_latency_ms"`
		MinLatencyMs int    `gorm:"column:min_latency_ms"`
		MaxLatencyMs int    `gorm:"column:max_latency_ms"`
		Count2xx     int    `gorm:"column:count_2xx"`
		Count3xx     int    `gorm:"column:count_3xx"`
		Count4xx     int    `gorm:"column:count_4xx"`
		Count5xx     int    `gorm:"column:count_5xx"`
		CountError   int    `gorm:"column:count_error"`
	}

	query := `
		SELECT 
			monitor_id,
			DATE(checked_at) AS date,
			COUNT(*) AS total_checks,
			COUNT(CASE WHEN status = 'up' THEN 1 END) AS up_checks,
			COUNT(CASE WHEN status = 'down' THEN 1 END) AS down_checks,
			ROUND(AVG(response_time_ms)) AS avg_latency_ms,
			COALESCE(MIN(CASE WHEN response_time_ms > 0 THEN response_time_ms END), 0) AS min_latency_ms,
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

	var deletedRows int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var results []AggregateResult
		if err := tx.Raw(query, threshold).Scan(&results).Error; err != nil {
			return err
		}

		if len(results) == 0 {
			return nil
		}

		log.Printf("[Retention] Aggregating %d daily stat bucket(s)...", len(results))

		for _, agg := range results {
			monitorUUID, err := uuid.Parse(agg.MonitorID)
			if err != nil {
				continue
			}

			if len(agg.Date) < 10 {
				continue
			}
			parsedDate, err := time.Parse("2006-01-02", agg.Date[:10])
			if err != nil {
				continue
			}

			uptimePct := 100.0
			if agg.TotalChecks > 0 {
				uptimePct = (float64(agg.UpChecks) / float64(agg.TotalChecks)) * 100.0
			}

			stat := model.MonitorDailyStat{
				MonitorID:        monitorUUID,
				Date:             parsedDate,
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

			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "monitor_id"}, {Name: "date"}},
				DoUpdates: clause.Assignments(map[string]interface{}{
					"total_checks":      gorm.Expr("monitor_daily_stats.total_checks + EXCLUDED.total_checks"),
					"up_checks":         gorm.Expr("monitor_daily_stats.up_checks + EXCLUDED.up_checks"),
					"down_checks":       gorm.Expr("monitor_daily_stats.down_checks + EXCLUDED.down_checks"),
					"uptime_percentage": gorm.Expr("CASE WHEN (monitor_daily_stats.total_checks + EXCLUDED.total_checks) > 0 THEN ROUND(((monitor_daily_stats.up_checks + EXCLUDED.up_checks) * 100.0) / (monitor_daily_stats.total_checks + EXCLUDED.total_checks), 2) ELSE 100.00 END"),
					"avg_latency_ms":    gorm.Expr("CASE WHEN (monitor_daily_stats.total_checks + EXCLUDED.total_checks) > 0 THEN ROUND((((monitor_daily_stats.avg_latency_ms * monitor_daily_stats.total_checks) + (EXCLUDED.avg_latency_ms * EXCLUDED.total_checks)) * 100.0) / ((monitor_daily_stats.total_checks + EXCLUDED.total_checks) * 100.0)) ELSE 0 END"),
					"min_latency_ms":    gorm.Expr("CASE WHEN monitor_daily_stats.min_latency_ms <= 0 THEN EXCLUDED.min_latency_ms WHEN EXCLUDED.min_latency_ms <= 0 THEN monitor_daily_stats.min_latency_ms ELSE CASE WHEN monitor_daily_stats.min_latency_ms < EXCLUDED.min_latency_ms THEN monitor_daily_stats.min_latency_ms ELSE EXCLUDED.min_latency_ms END END"),
					"max_latency_ms":    gorm.Expr("CASE WHEN monitor_daily_stats.max_latency_ms > EXCLUDED.max_latency_ms THEN monitor_daily_stats.max_latency_ms ELSE EXCLUDED.max_latency_ms END"),
					"count2xx":          gorm.Expr("monitor_daily_stats.count2xx + EXCLUDED.count2xx"),
					"count3xx":          gorm.Expr("monitor_daily_stats.count3xx + EXCLUDED.count3xx"),
					"count4xx":          gorm.Expr("monitor_daily_stats.count4xx + EXCLUDED.count4xx"),
					"count5xx":          gorm.Expr("monitor_daily_stats.count5xx + EXCLUDED.count5xx"),
					"count_error":       gorm.Expr("monitor_daily_stats.count_error + EXCLUDED.count_error"),
					"updated_at":        time.Now().UTC(),
				}),
			}).Create(&stat).Error; err != nil {
				log.Printf("[Retention] Failed to upsert daily stat for monitor %s on date %s: %v", agg.MonitorID, parsedDate.Format("2006-01-02"), err)
				return err
			}
		}

		result := tx.Where("checked_at < ?", threshold).Delete(&model.Check{})
		if result.Error != nil {
			return result.Error
		}
		deletedRows = result.RowsAffected
		return nil
	})
	if err != nil {
		log.Printf("[Retention] Aborted cleanup to prevent data loss: %v", err)
		return
	}

	log.Printf("[Retention] Success clean old data check: %d row(s)", deletedRows)
}

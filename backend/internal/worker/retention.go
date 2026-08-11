package worker

import (
	"PulseWatch/internal/model"
	"log"
	"time"

	"gorm.io/gorm"
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
	threshold := time.Now().Add(-3 * 24 * time.Hour)

	log.Printf("[Retention] Cleaning old data check before : %s", threshold.Format("2006-01-02 15:04:05"))

	result := r.db.Where("checked_at < ?", threshold).Delete(&model.Check{})

	if result.Error != nil {
		log.Printf("[Retention] Failed to clean old data check: %v", result.Error)
		return
	}

	log.Printf("[Retention] Success clean old data check: %d row(s)", result.RowsAffected)
}

package worker

import (
	"netra-uptime/internal/monitor"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Scheduler struct {
	repo       monitor.MonitorRepository
	workerpool *WorkerPool
	stopChan   chan struct{}
	lastRun    map[uuid.UUID]time.Time
	mu         sync.Mutex
}

func NewScheduler(repo monitor.MonitorRepository, wp *WorkerPool) *Scheduler {
	return &Scheduler{
		repo:       repo,
		workerpool: wp,
		stopChan:   make(chan struct{}),
		lastRun:    make(map[uuid.UUID]time.Time),
		mu:         sync.Mutex{},
	}
}

func (s *Scheduler) Start() {
	ticker := time.NewTicker(20 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.scheduleActiveMonitors()
			case <-s.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	close(s.stopChan)
}

func (s *Scheduler) scheduleActiveMonitors() {
	monitors, err := s.repo.GetActiveMonitor()
	if err != nil {
		log.Printf("Scheduler failed to get active monitor: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, monitor := range monitors {
		lastTime, exist := s.lastRun[monitor.ID]
		interval := time.Duration(monitor.IntervalSeconds) * time.Second

		now := time.Now()

		if !exist || now.Sub(lastTime) >= interval {
			s.workerpool.QueueJob(Job{Monitor: monitor})
			s.lastRun[monitor.ID] = now
		}
	}
}

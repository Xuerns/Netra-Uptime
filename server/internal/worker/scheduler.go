package worker

import (
	"hash/crc32"
	"log"
	"netra-uptime/internal/model"
	"netra-uptime/internal/monitor"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Scheduler struct {
	repo       monitor.MonitorRepository
	workerpool *WorkerPool
	stopChan   chan struct{}
	lastRun    map[uuid.UUID]time.Time
	nextRun    map[uuid.UUID]time.Time
	mu         sync.Mutex
}

func NewScheduler(repo monitor.MonitorRepository, wp *WorkerPool) *Scheduler {
	return &Scheduler{
		repo:       repo,
		workerpool: wp,
		stopChan:   make(chan struct{}),
		lastRun:    make(map[uuid.UUID]time.Time),
		nextRun:    make(map[uuid.UUID]time.Time),
		mu:         sync.Mutex{},
	}
}

func (s *Scheduler) Start() {
	ticker := time.NewTicker(5 * time.Second)
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

func (s *Scheduler) ScheduleMonitor(mon model.Monitor) {
	s.mu.Lock()
	defer s.mu.Unlock()

	intervalSec := mon.IntervalSeconds
	if intervalSec <= 0 {
		intervalSec = 60
	}
	initialDelay := time.Duration(crc32.ChecksumIEEE([]byte(mon.ID.String()))%uint32(intervalSec)) * time.Second
	nextRun := time.Now().Add(initialDelay)
	s.nextRun[mon.ID] = nextRun
}

func (s *Scheduler) scheduleActiveMonitors() {
	monitors, err := s.repo.GetActiveMonitor()
	if err != nil {
		log.Printf("Scheduler failed to get active monitor: %v", err)
		return
	}

	var jobsToQueue []Job

	s.mu.Lock()
	now := time.Now()
	for _, mon := range monitors {
		intervalSec := mon.IntervalSeconds
		if intervalSec <= 0 {
			intervalSec = 60
		}
		targetNextRun, exist := s.nextRun[mon.ID]
		if !exist {
			initialDelay := time.Duration(crc32.ChecksumIEEE([]byte(mon.ID.String()))%uint32(intervalSec)) * time.Second
			targetNextRun = now.Add(initialDelay)
			s.nextRun[mon.ID] = targetNextRun
			continue
		}

		if now.After(targetNextRun) || now.Equal(targetNextRun) {
			jobsToQueue = append(jobsToQueue, Job{Monitor: mon})
			s.lastRun[mon.ID] = now
			s.nextRun[mon.ID] = now.Add(time.Duration(intervalSec) * time.Second)
		}
	}
	s.mu.Unlock()

	for _, job := range jobsToQueue {
		s.workerpool.QueueJob(job)
	}
}

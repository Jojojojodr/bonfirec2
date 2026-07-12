package bonfirec2

import (
	"errors"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

var scheduler = &TaskScheduler{interval: 2 * time.Second}

type TaskScheduler struct {
	db       *gorm.DB
	interval time.Duration
	mu       sync.Mutex
	running  bool
	stopCh   chan struct{}
	doneCh   chan struct{}
}

func (s *TaskScheduler) Start(db *gorm.DB) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.running {
		return nil
	}
	if db == nil {
		return errors.New("database is required")
	}
	
	s.db = db
	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})
	s.running = true
	
	go s.loop(s.stopCh, s.doneCh)
	return nil
}

func (s *TaskScheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	stopCh := s.stopCh
	doneCh := s.doneCh
	s.running = false
	s.mu.Unlock()
	
	close(stopCh)
	<-doneCh
}

func (s *TaskScheduler) loop(stopCh <-chan struct{}, doneCh chan<- struct{}) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	defer close(doneCh)
	
	s.processDueTasks()
	
	for {
		select {
		case <-ticker.C:
			s.processDueTasks()
		case <-stopCh:
			return
		}
	}
}

func (s *TaskScheduler) processDueTasks() {
	if s.db == nil {
		return
	}
	
	now := currentTaskTimestamp()
	var tasks []Task
	if err := s.db.Where("status IN ? AND next_run_at <= ?", []string{"scheduled", "waiting"}, now).
		Order("next_run_at ASC, created_at ASC").Find(&tasks).Error; err != nil {
		log.Printf("failed to load due tasks: %v", err)
		return
	}
	
	for i := range tasks {
		s.dispatchTask(&tasks[i])
	}
}

func (s *TaskScheduler) dispatchTask(task *Task) {
	grunt := Grunts[task.GruntID]
	if grunt == nil {
		task.markRetry("grunt not connected")
		if err := LogTaskWaiting(task, "grunt not connected"); err != nil {
			log.Printf("failed to save task %s waiting log: %v", task.ID, err)
		}
		NotifyTaskWaiting(task, "grunt not connected")
		if err := s.db.Save(task).Error; err != nil {
			log.Printf("failed to update task %s while waiting for grunt: %v", task.ID, err)
		}
		return
	}
	
	if err := SendCommandToGrunt(task.GruntID, task.Command); err != nil {
		task.markRetry(err.Error())
		if logErr := LogTaskWaiting(task, err.Error()); logErr != nil {
			log.Printf("failed to save task %s retry log: %v", task.ID, logErr)
		}
		NotifyTaskWaiting(task, err.Error())
		if err := s.db.Save(task).Error; err != nil {
			log.Printf("failed to mark task %s for retry: %v", task.ID, err)
		}
		return
	}
	
	if err := SaveGruntMessage(grunt.ListenerID, grunt.ID, task.Command, true); err != nil {
		log.Printf("failed to save task %s dispatch message: %v", task.ID, err)
	}
	
	task.markDispatched()
	if err := LogTaskDispatched(task); err != nil {
		log.Printf("failed to save task %s dispatch log: %v", task.ID, err)
	}
	NotifyTaskDispatched(task)
	if err := s.db.Save(task).Error; err != nil {
		log.Printf("failed to persist task %s dispatch state: %v", task.ID, err)
	}
}

func StartTaskScheduler(db *gorm.DB) error {
	return scheduler.Start(db)
}

func StopTaskScheduler() {
	scheduler.Stop()
}
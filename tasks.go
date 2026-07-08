package bonfirec2

import (
	"errors"
	"strings"
	"time"

	"github.com/Jojojojodr/bonfirec2/models"
	"github.com/google/uuid"
)

type Task struct {
	models.DefaultModel
	GruntID            string `json:"grunt_id"`
	Command            string `json:"command"`
	Status             string `json:"status"`
	Result             string `json:"result"`
	Repeat             bool   `json:"repeat"`
	RepeatEverySeconds int    `json:"repeat_every_seconds"`
	RepeatCount        int    `json:"repeat_count"`
	RunCount           int    `json:"run_count"`
	ScheduledFor       string `json:"scheduled_for"`
	NextRunAt          string `json:"next_run_at"`
	LastRunAt          string `json:"last_run_at"`
	CompletedAt        string `json:"completed_at"`
	CancelledAt        string `json:"cancelled_at"`
	LastError          string `json:"last_error"`
	Timeout            int    `json:"timeout"`
}

func (t *Task) Cancel() error {
	if t == nil {
		return errors.New("task is nil")
	}
	
	now := currentTaskTimestamp()
	t.Status = "cancelled"
	t.CancelledAt = now
	t.UpdatedAt = now
	
	db := Data.GetDB()
	return db.Save(t).Error
}

func (t *Task) markDispatched() {
	now := currentTaskTimestamp()
	t.RunCount++
	t.LastRunAt = now
	t.LastError = ""
	t.UpdatedAt = now
	if !t.Repeat || (t.RepeatCount > 0 && t.RunCount >= t.RepeatCount) {
		t.Status = "completed"
		t.CompletedAt = now
		t.NextRunAt = ""
		return
	}
	
	t.Status = "scheduled"
	if t.RepeatEverySeconds > 0 {
		t.NextRunAt = time.Now().Add(time.Duration(t.RepeatEverySeconds) * time.Second).Format("2006-01-02 15:04:05")
	}
}

func (t *Task) markRetry(reason string) {
	now := currentTaskTimestamp()
	t.Status = "waiting"
	t.LastError = reason
	t.UpdatedAt = now
}

func currentTaskTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func NewTask(gruntID, command string, scheduledFor time.Time, repeat bool, repeatEverySeconds, repeatCount, timeout int) (*Task, error) {
	gruntID = strings.TrimSpace(gruntID)
	command = strings.TrimSpace(command)
	if gruntID == "" {
		return nil, errors.New("grunt_id is required")
	}
	if command == "" {
		return nil, errors.New("command is required")
	}
	if repeat && repeatEverySeconds <= 0 {
		return nil, errors.New("repeat_every_seconds must be greater than zero when repeat is enabled")
	}
	if repeatCount < 0 {
		return nil, errors.New("repeat_count cannot be negative")
	}

	now := currentTaskTimestamp()
	task := &Task{
		DefaultModel: models.DefaultModel{
			ID:        uuid.New().String(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		GruntID:            gruntID,
		Command:            command,
		Status:             "scheduled",
		Repeat:             repeat,
		RepeatEverySeconds: repeatEverySeconds,
		RepeatCount:        repeatCount,
		ScheduledFor:       scheduledFor.Format("2006-01-02 15:04:05"),
		NextRunAt:          scheduledFor.Format("2006-01-02 15:04:05"),
		Timeout:            timeout,
	}

	if !repeat {
		task.RepeatCount = 1
	}

	db := Data.GetDB()
	if err := db.Create(task).Error; err != nil {
		return nil, err
	}

	return task, nil
}

func GetTasks() ([]*Task, error) {
	db := Data.GetDB()
	var tasks []*Task
	if err := db.Order("next_run_at ASC, created_at DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func GetTaskByID(id string) (*Task, error) {
	db := Data.GetDB()
	var task Task
	if err := db.First(&task, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}
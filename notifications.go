package bonfirec2

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Jojojojodr/bonfirec2/models"

	"github.com/google/uuid"
)

var Notifications = NewNotificationCenter(200)

type Notification struct {
	models.DefaultModel
	EventType string            `json:"event_type"`
	Level     string            `json:"level"`
	Title     string            `json:"title"`
	Message   string            `json:"message"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type NotificationCenter struct {
	mu    sync.RWMutex
	max   int
	items []Notification
}

func (c *NotificationCenter) Publish(eventType, level, title, message string, metadata map[string]string) Notification {
	notification := Notification{
		DefaultModel: models.DefaultModel{
			ID: uuid.New().String(),
			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
			UpdatedAt: time.Now().Format("2006-01-02 15:04:05"),
		},
		EventType: strings.TrimSpace(eventType),
		Level:     strings.TrimSpace(level),
		Title:     strings.TrimSpace(title),
		Message:   strings.TrimSpace(message),
		Metadata:  metadata,
	}
	
	c.mu.Lock()
	c.items = append(c.items, notification)
	if len(c.items) > c.max {
		c.items = c.items[len(c.items)-c.max:]
	}
	c.mu.Unlock()
	
	return notification
}

func (c *NotificationCenter) Latest(limit int) []Notification {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	count := len(c.items)
	if count == 0 {
		return []Notification{}
	}
	
	if limit <= 0 || limit > count {
		limit = count
	}
	
	result := make([]Notification, 0, limit)
	for i := count - 1; i >= count-limit; i-- {
		result = append(result, c.items[i])
	}
	
	return result
}

func NewNotificationCenter(max int) *NotificationCenter {
	if max <= 0 {
		max = 200
	}
	return &NotificationCenter{max: max}
}

func NotifyGruntConnected(gruntID, listenerID, address string) {
	Notifications.Publish(
		"grunt_connected",
		"success",
		"Grunt connected",
		fmt.Sprintf("Grunt %s connected from %s on listener %s", gruntID, address, listenerID),
		map[string]string{
			"grunt_id":    gruntID,
			"listener_id": listenerID,
			"address":     address,
		},
	)
}

func NotifyGruntDisconnected(gruntID, listenerID, address, reason string) {
	if strings.TrimSpace(reason) == "" {
		reason = "connection closed"
	}

	Notifications.Publish(
		"grunt_disconnected",
		"warning",
		"Grunt disconnected",
		fmt.Sprintf("Grunt %s disconnected from %s (%s)", gruntID, address, reason),
		map[string]string{
			"grunt_id":    gruntID,
			"listener_id": listenerID,
			"address":     address,
			"reason":      reason,
		},
	)
}

func NotifyGruntMessage(gruntID, listenerID, content string) {
	trimmed := strings.TrimSpace(content)
	if len(trimmed) > 160 {
		trimmed = trimmed[:160] + "..."
	}

	Notifications.Publish(
		"grunt_message",
		"info",
		"Message from grunt",
		fmt.Sprintf("%s: %s", gruntID, trimmed),
		map[string]string{
			"grunt_id":    gruntID,
			"listener_id": listenerID,
		},
	)
}

func NotifyTaskDispatched(task *Task) {
	if task == nil {
		return
	}

	level := "info"
	if task.Status == "completed" {
		level = "success"
	}

	Notifications.Publish(
		"task_dispatched",
		level,
		"Task executed",
		fmt.Sprintf("Task %s sent command '%s' to grunt %s (run %d)", task.ID, task.Command, task.GruntID, task.RunCount),
		map[string]string{
			"task_id":   task.ID,
			"grunt_id":  task.GruntID,
			"status":    task.Status,
			"run_count": fmt.Sprintf("%d", task.RunCount),
		},
	)
}

func NotifyTaskWaiting(task *Task, reason string) {
	if task == nil {
		return
	}

	Notifications.Publish(
		"task_waiting",
		"warning",
		"Task waiting",
		fmt.Sprintf("Task %s for grunt %s is waiting: %s", task.ID, task.GruntID, reason),
		map[string]string{
			"task_id":  task.ID,
			"grunt_id": task.GruntID,
			"reason":   reason,
		},
	)
}

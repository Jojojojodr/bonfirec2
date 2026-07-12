package bonfirec2

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Jojojojodr/bonfirec2/models"
	"github.com/google/uuid"
)

type EventLog struct {
	models.DefaultModel
	EventType  string `json:"event_type"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	ListenerID string `json:"listener_id,omitempty"`
	GruntID    string `json:"grunt_id,omitempty"`
	TaskID     string `json:"task_id,omitempty"`
}

func GetEventLogs(limit int, gruntID, taskID string) ([]*EventLog, error) {
	if Data == nil {
		return nil, nil
	}
	
	db := Data.GetDB()
	if db == nil {
		return nil, nil
	}
	
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	
	query := db.Model(&EventLog{})
	if strings.TrimSpace(gruntID) != "" {
		query = query.Where("grunt_id = ?", strings.TrimSpace(gruntID))
	}
	if strings.TrimSpace(taskID) != "" {
		query = query.Where("task_id = ?", strings.TrimSpace(taskID))
	}
	
	var logs []*EventLog
	if err := query.Order("created_at DESC").Limit(limit).Find(&logs).Error; err != nil {
		log.Printf("Failed to retrieve event logs from database: %v", err)
		return nil, err
	}
	
	return logs, nil
}

func LogGruntConnected(gruntID, listenerID, address string) error {
	return logEvent(
		"grunt_connected",
		"success",
		fmt.Sprintf("Grunt %s connected from %s on listener %s", gruntID, address, listenerID),
		listenerID,
		gruntID,
		"",
	)
}

func LogGruntDisconnected(gruntID, listenerID, address, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "connection closed"
	}
	
	return logEvent(
		"grunt_disconnected",
		"warning",
		fmt.Sprintf("Grunt %s disconnected from %s (%s)", gruntID, address, reason),
		listenerID,
		gruntID,
		"",
	)
}

func LogGruntMessageSent(listenerID, gruntID, content string) error {
	trimmed := strings.TrimSpace(content)
	if len(trimmed) > 160 {
		trimmed = trimmed[:160] + "..."
	}
	
	return logEvent(
		"grunt_message_sent",
		"info",
		fmt.Sprintf("Sent to grunt %s: %s", gruntID, trimmed),
		listenerID,
		gruntID,
		"",
	)
}

func LogTaskDispatched(task *Task) error {
	if task == nil {
		return nil
	}
	
	level := "info"
	if task.Status == "completed" {
		level = "success"
	}
	
	return logEvent(
		"task_dispatched",
		level,
		fmt.Sprintf("Task %s sent command '%s' to grunt %s (run %d)", task.ID, task.Command, task.GruntID, task.RunCount),
		"",
		task.GruntID,
		task.ID,
	)
}

func LogTaskWaiting(task *Task, reason string) error {
	if task == nil {
		return nil
	}
	
	return logEvent(
		"task_waiting",
		"warning",
		fmt.Sprintf("Task %s for grunt %s is waiting: %s", task.ID, task.GruntID, reason),
		"",
		task.GruntID,
		task.ID,
	)
}

func logEvent(eventType, severity, message, listenerID, gruntID, taskID string) error {
	if Data == nil {
		return nil
	}

	db := Data.GetDB()
	if db == nil {
		return nil
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	entry := &EventLog{
		DefaultModel: models.DefaultModel{
			ID:        uuid.New().String(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		EventType:  strings.TrimSpace(eventType),
		Severity:   strings.TrimSpace(severity),
		Message:    strings.TrimSpace(message),
		ListenerID: strings.TrimSpace(listenerID),
		GruntID:    strings.TrimSpace(gruntID),
		TaskID:     strings.TrimSpace(taskID),
	}

	if err := db.Create(entry).Error; err != nil {
		return err
	}

	return nil
}

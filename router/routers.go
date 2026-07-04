package router

import (
	"github.com/Jojojojodr/bonfirec2/controller"
	"github.com/gin-gonic/gin"
)

func SetupRouter(router *gin.Engine) *gin.Engine {
	router.Static("/static", "./static")

	router.GET("/", controller.HomeView)

	listeners := router.Group("/listeners")
	listeners.GET("/", controller.ListenersView)
	listeners.GET("/l", controller.ListenerDetailView)

	grunts := router.Group("/grunts")
	grunts.GET("/", controller.GruntsView)
	grunts.GET("/g", controller.GruntDetailView)

	router.NoRoute(controller.NotFound)

	return router
}

func SetupApiRouter(router *gin.Engine) *gin.Engine {
	api := router.Group("/api")
	api.GET("/health", controller.GetHealth)
	api.GET("/grunts", controller.GetGrunts)

	tasks := api.Group("/tasks")
	tasks.GET("/", controller.GetTasks)
	tasks.POST("/", controller.CreateTask)

	listeners := api.Group("/listeners")
	listeners.GET("/", controller.GetListeners)
	listeners.POST("/", controller.CreateListener)

	messages := api.Group("/messages")
	messages.GET("/", controller.GetLatestMessages)
	messages.GET("/grunt", controller.GetGruntMessages)
	messages.POST("/grunt", controller.SendGruntMessage)

	return router
}

func SetupActionRouter(router *gin.Engine) *gin.Engine {
	action := router.Group("/actions")
	action.POST("/new-listener", controller.NewListener)
	action.POST("/start-listener", controller.StartListener)
	action.POST("/stop-listener", controller.StopListener)
	action.POST("/grunts/terminal/command", controller.SendGruntCommand)

	partials := action.Group("/partials")
	partials.GET("/dashboard/active-grunts", controller.DashboardActiveGruntsPartial)
	partials.GET("/grunts/terminal/messages", controller.GruntTerminalMessagesPartial)

	return router
}

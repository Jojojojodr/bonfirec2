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
	api.GET("/notifications", controller.GetNotifications)
	api.GET("/messages", controller.GetLatestMessages)

	tasks := api.Group("/tasks")
	tasks.GET("/", controller.GetTasks)
	tasks.POST("/", controller.CreateTask)

	grunts := api.Group("/grunts")
	grunts.GET("/", controller.GetGrunts)

	grunt := grunts.Group("/g")
	grunt.GET("/", controller.GetGruntById)
	
	messages := grunt.Group("/messages")
	messages.GET("/", controller.GetGruntMessages)
	messages.POST("/m", controller.SendGruntMessage)

	listeners := api.Group("/listeners")
	listeners.GET("/", controller.GetListeners)
	listeners.POST("/", controller.CreateListener)

	listener := listeners.Group("/l")
	listener.GET("/", controller.GetListenerById)
	listener.POST("/start", controller.StartListener)
	listener.POST("/stop", controller.StopListener)

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
	partials.GET("/dashboard/notifications", controller.DashboardNotificationsPartial)
	partials.GET("/grunts/terminal/messages", controller.GruntTerminalMessagesPartial)

	return router
}

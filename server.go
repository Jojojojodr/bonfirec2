package bonfirec2

import (
	"log"

	"github.com/Jojojojodr/bonfirec2/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	Engine *gin.Engine
}

func (s *Server) Start() {
	log.Println("Starting server on port ", config.AppConfig.Server.Port)
	log.Println("Go to http://localhost:" + config.AppConfig.Server.Port + " to access the server.")
	if err := s.Engine.Run(":" + config.AppConfig.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func NewServer() *Server {
	// Set Gin mode based on configuration
	if config.AppConfig.Server.Debug {
		log.Println("Running in debug mode")
		gin.SetMode(gin.DebugMode)
	} else {
		log.Println("Running in release mode")
		gin.SetMode(gin.ReleaseMode)
	}
	
	// Create a new Gin engine
	engine := gin.Default()
	
	// Configure CORS middleware
	engine.Use(cors.Default())


	err := engine.SetTrustedProxies(nil)
	if err != nil {
		log.Fatalf("Failed to set trusted proxies: %v", err)
	}

	return &Server{
		Engine: engine,
	}
}

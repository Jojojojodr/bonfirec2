package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Jojojojodr/bonfirec2"
	"github.com/Jojojojodr/bonfirec2/config"
	"github.com/Jojojojodr/bonfirec2/pkg/app"
	"github.com/Jojojojodr/bonfirec2/router"
)

func main() {
	// Load configuration
	config.LoadConfig()

	// Initialize database
	db := bonfirec2.NewDatabase()
	if err := db.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Create server
	log.Println("Starting Server...")
	server := bonfirec2.NewServer()

	// Setup routes
	log.Println("Setting up routes...")
	router.SetupRouter(server.Engine)
	router.SetupApiRouter(server.Engine)
	router.SetupActionRouter(server.Engine)

	// Safety net for dev reloads (e.g. Air): clear stale active sessions on boot.
	// Set debug mode in config.yaml to true to enable this behavior.
	if config.AppConfig.Server.Debug {
		log.Println("Debug mode enabled: clearing all active grunts on boot...")
		bonfirec2.SetAllGruntsInactive()
	}

	// Restart services
	log.Println("Restarting services...")
	app.RestartServices(db.GetDB())

	// Graceful shutdown hook for Ctrl+C and termination signals.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go app.ShutdownServices(sigCh)

	// Start server
	server.Start()
}
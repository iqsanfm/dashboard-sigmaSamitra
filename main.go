package main

import (
	"log"
	"net/http"

	"github.com/iqsanfm/dashboard-pekerjaan-backend/config"
	"github.com/iqsanfm/dashboard-pekerjaan-backend/internal/routes" // Import routes
	"github.com/iqsanfm/dashboard-pekerjaan-backend/pkg/database"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database connection
	db, err := database.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close() // Ensure the database connection is closed when main exits

	// Setup Gin router
	router := routes.SetupRouter(db, cfg) // Pass the database connection to setup routes

	// Start the HTTP server
	log.Printf("Server starting on port %s", cfg.ServerPort)
	err = http.ListenAndServe(cfg.ServerPort, router)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
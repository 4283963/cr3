package main

import (
	"escape-room/config"
	"escape-room/middleware"
	"escape-room/models"
	"escape-room/routes"
	"log"
)

func main() {
	cfg := config.Load()

	models.InitDB(cfg)

	r := routes.SetupRouter()

	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := r.Run(cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

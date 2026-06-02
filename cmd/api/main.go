package main

import (
	"log"

	"github.com/Fluorineheit/duwit-tracker-be/internal/config"
	"github.com/Fluorineheit/duwit-tracker-be/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg := config.Load()

	router := server.NewRouter(cfg)

	address := ":" + cfg.AppPort

	log.Printf("Starting %s on http://localhost%s", cfg.AppName, address)

	if err := router.Run(address); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
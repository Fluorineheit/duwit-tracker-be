package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Fluorineheit/duwit-tracker-be/internal/config"
	"github.com/Fluorineheit/duwit-tracker-be/internal/database"
	"github.com/Fluorineheit/duwit-tracker-be/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.NewPostgresPool(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	router := server.NewRouter(cfg, db)

	address := ":" + cfg.AppPort

	log.Printf("Starting %s on http://localhost%s", cfg.AppName, address)

	if err := router.Run(address); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

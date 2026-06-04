package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

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
	defer db.Close()

	router := server.NewRouter(cfg, db)

	listener, address, err := listen(ctx, cfg.AppEnv, cfg.AppPort)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	httpServer := &http.Server{
		Addr:    address,
		Handler: router,
	}

	log.Printf("Starting %s on http://localhost%s", cfg.AppName, address)

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.Serve(listener)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		log.Println("Shutting down server...")
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("Failed to shutdown server: %v", err)
		}
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}
}

func listen(ctx context.Context, appEnv string, appPort string) (net.Listener, string, error) {
	address := ":" + appPort
	listenConfig := net.ListenConfig{}

	listener, err := listenConfig.Listen(ctx, "tcp", address)
	if err == nil {
		return listener, address, nil
	}

	if appEnv == "production" || !isAddressInUse(err) {
		return nil, "", err
	}

	port, parseErr := strconv.Atoi(appPort)
	if parseErr != nil {
		return nil, "", err
	}

	for nextPort := port + 1; nextPort <= port+20; nextPort++ {
		nextAddress := ":" + strconv.Itoa(nextPort)

		listener, listenErr := listenConfig.Listen(ctx, "tcp", nextAddress)
		if listenErr == nil {
			log.Printf("Port %s is already in use. Falling back to http://localhost%s", appPort, nextAddress)
			return listener, nextAddress, nil
		}

		if !isAddressInUse(listenErr) {
			return nil, "", listenErr
		}
	}

	return nil, "", err
}

func isAddressInUse(err error) bool {
	if errors.Is(err, syscall.EADDRINUSE) {
		return true
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "address already in use") ||
		strings.Contains(message, "only one usage of each socket address")
}

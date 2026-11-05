package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"netra-uptime/config"
	"netra-uptime/internal/app"
)

func main() {
	cfg := config.Load()
	serverApp, err := app.Setup(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Netra Uptime: %v", err)
	}

	log.Printf("[Netra Uptime] Server running at http://%s:%s (DB Driver: %s)\n", cfg.Host, cfg.Port, cfg.DBDriver)

	go func() {
		if err := serverApp.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to run on %s: %v", serverApp.Server.Addr, err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("[Netra Uptime] Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := serverApp.Shutdown(ctx); err != nil {
		log.Printf("[Netra Uptime] Shutdown warning: %v\n", err)
	} else {
		log.Println("[Netra Uptime] Server exited cleanly")
	}
}

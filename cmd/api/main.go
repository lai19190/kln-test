package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kln-test/internal/config"
	"kln-test/internal/handlers"
	"kln-test/internal/holidays"
	"kln-test/internal/middleware"
)

const (
	serverPort = "8080"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create middleware chain
	middlewareChain := middleware.Chain(
		middleware.ConfigReload(cfg),
		middleware.Auth(cfg),
		middleware.ValidateHeaders(),
		middleware.Recovery(),
	)

	// Initialize handlers
	subscriptionHandler := handlers.NewSubscriptionHandler(cfg)
	holidaysHandler := handlers.NewHolidaysFetchHandler(holidays.NewService(holidays.NewClient()))

	// Setup router
	mux := http.NewServeMux()
	mux.Handle("/subscriptions", middlewareChain(subscriptionHandler))
	mux.Handle("/public-holidays", middlewareChain(holidaysHandler))

	// Create server
	srv := &http.Server{
		Addr:    ":" + serverPort,
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", serverPort)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
}

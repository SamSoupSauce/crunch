package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"crunch/server/internal/api"
	"crunch/server/internal/room"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	store := room.NewStore()

	// Periodic janitor to purge expired rooms (>24h)
	stopJanitor := make(chan struct{})
	store.StartJanitor(10*time.Minute, stopJanitor)

	server := api.NewServer(store)
	router := api.NewRouter(server)

	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown channel
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("💥 Crunch Multiplayer API Server listening on :%s", port)
		log.Printf("   • POST /rooms               -> Create a 24h room")
		log.Printf("   • POST /rooms/{id}/join     -> Join room (player 1 / player 2)")
		log.Printf("   • POST /rooms/{id}/turn     -> Take a turn (strict player turn order)")
		log.Printf("   • POST /rooms/{id}/messages -> Send message")
		log.Printf("   • GET  /rooms/{id}/messages -> Retrieve messages")
		log.Printf("   • GET  /rooms/{id}          -> Room status")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")

	close(stopJanitor)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	log.Println("Server gracefully stopped.")
}

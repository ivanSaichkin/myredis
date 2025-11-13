package main

import (
	"ivanSaichkin/myredis/internal/config"
	"ivanSaichkin/myredis/internal/server"
	"ivanSaichkin/myredis/internal/storage"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.DefaulteConfig()
	store := storage.NewMemoryStorageWithPersistence(&cfg.Persistence)

	if err := store.StartPersistence(); err != nil {
		log.Printf("Warning: failed to start persistence: %v", err)
	}

	store.StartExpirationChecker(30 * time.Second)

	handler := server.NewHandler(store)
	server := server.NewTCPServer(cfg.Address, handler)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Starting MyRedis server...")
		log.Printf("Server listening on %s", cfg.Address)
		log.Printf("Persistence enabled: %v", cfg.Persistence.Enabled)
		if err := server.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	sig := <-sigChan
	log.Printf("Received signal %v, shutting down...", sig)
	if cfg.Persistence.Enabled {
		log.Println("Stopping persistence...")
		if err := store.StopPersistence(); err != nil {
			log.Printf("Error stopping persistence: %v", err)
		} else {
			log.Println("Persistence stopped successfully")
		}
	}
	log.Println("MyRedis server stopped")
}

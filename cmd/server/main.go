package main

import (
	"ivanSaichkin/myredis/internal/server"
	"ivanSaichkin/myredis/internal/storage"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	store := storage.NewMemoryStorage()

	store.StartExpirationChecker(30 * time.Second)

	handler := server.NewHandler(store)
	server := server.NewTCPServer(":6379", handler)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Starting MyRedis server...")
		if err := server.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	sig := <-sigChan
	log.Printf("Received signal %v, shutting down...", sig)
	log.Println("MyRedis server stopped")
}

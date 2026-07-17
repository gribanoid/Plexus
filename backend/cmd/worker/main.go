package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/plexus/backend/internal/config"
	"github.com/plexus/backend/internal/delivery"
)

// plexus-worker runs asynq jobs (email, search index, webhooks) without serving HTTP.
// Use the same Docker image with command ["./plexus-worker"] and set RUN_WORKERS=false on API replicas.
func main() {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	rt, err := delivery.NewRuntime(cfg)
	if err != nil {
		log.Fatalf("runtime: %v", err)
	}
	defer rt.Close()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("plexus worker starting")
		if err := rt.JobServer.Start(); err != nil {
			log.Fatalf("job server error: %v", err)
		}
	}()

	<-quit
	log.Println("worker shutting down...")
	rt.JobServer.Shutdown()
}

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/plexus/backend/internal/api"
	"github.com/plexus/backend/internal/config"
	"github.com/plexus/backend/internal/delivery"
)

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

	go rt.Hub.Run()

	if cfg.RunWorkers {
		go func() {
			if err := rt.JobServer.Start(); err != nil {
				log.Printf("job server error: %v", err)
			}
		}()
	}

	app := api.New(api.Dependencies{
		Config:       cfg,
		Pool:         rt.Pool,
		Redis:        rt.Redis,
		SearchClient: rt.Search,
		Hub:          rt.Hub,
		JobClient:    rt.JobClient,
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := ":" + cfg.Port
		log.Printf("plexus server listening on %s (workers=%v)", addr, cfg.RunWorkers)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down...")
	if cfg.RunWorkers {
		rt.JobServer.Shutdown()
	}
	if err := app.Shutdown(); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}

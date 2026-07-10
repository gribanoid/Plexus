package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/plexus/backend/internal/api"
	"github.com/plexus/backend/internal/config"
	"github.com/plexus/backend/internal/db"
	"github.com/plexus/backend/internal/jobs"
	"github.com/plexus/backend/internal/repository"
	"github.com/plexus/backend/internal/search"
	"github.com/plexus/backend/internal/websocket"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	pool, err := db.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	redisClient := db.NewRedis(cfg.RedisURL)
	defer redisClient.Close()

	searchClient := search.NewClient(cfg.MeilisearchURL, cfg.MeilisearchKey)
	if err := searchClient.EnsureIndexes(context.Background()); err != nil {
		log.Printf("meilisearch ensure indexes: %v", err)
	}

	repo := repository.New(pool)

	hub := websocket.NewHub(redisClient)
	go hub.Run()

	jobServer := jobs.NewServer(cfg.RedisURL, searchClient, repo)
	go func() {
		if err := jobServer.Start(); err != nil {
			log.Printf("job server error: %v", err)
		}
	}()

	app := api.New(api.Dependencies{
		Config:       cfg,
		Pool:         pool,
		Redis:        redisClient,
		SearchClient: searchClient,
		Hub:          hub,
		JobClient:    jobs.NewClient(cfg.RedisURL),
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := ":" + cfg.Port
		log.Printf("plexus server listening on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down...")
	jobServer.Shutdown()
	if err := app.Shutdown(); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}

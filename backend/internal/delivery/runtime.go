package delivery

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/plexus/backend/internal/config"
	"github.com/plexus/backend/internal/crypto"
	"github.com/plexus/backend/internal/db"
	"github.com/plexus/backend/internal/jobs"
	"github.com/plexus/backend/internal/repository"
	"github.com/plexus/backend/internal/search"
	"github.com/plexus/backend/internal/websocket"
	"github.com/redis/go-redis/v9"
)

// Runtime holds shared process dependencies for API and worker entrypoints.
type Runtime struct {
	Cfg       *config.Config
	Pool      *pgxpool.Pool
	Redis     *redis.Client
	Search    *search.Client
	Repo      *repository.Repo
	Hub       *websocket.Hub
	JobServer *jobs.Server
	JobClient *asynq.Client
}

// NewRuntime connects infra dependencies shared by API and worker processes.
func NewRuntime(cfg *config.Config) (*Runtime, error) {
	pool, err := db.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
		pool.Close()
		return nil, err
	}

	rdb := db.NewRedis(cfg.RedisURL)
	searchClient := search.NewClient(cfg.MeilisearchURL, cfg.MeilisearchKey)
	if err := searchClient.EnsureIndexes(context.Background()); err != nil {
		slog.Warn("meilisearch ensure indexes", "error", err)
	}

	repo := repository.New(pool)
	hub := websocket.NewHub(rdb)

	return &Runtime{
		Cfg:       cfg,
		Pool:      pool,
		Redis:     rdb,
		Search:    searchClient,
		Repo:      repo,
		Hub:       hub,
		JobServer: jobs.NewServer(cfg.RedisURL, searchClient, repo, crypto.KeyFromString(cfg.EncryptionKey)),
		JobClient: jobs.NewClient(cfg.RedisURL),
	}, nil
}

func (rt *Runtime) Close() {
	if rt.JobClient != nil {
		_ = rt.JobClient.Close()
	}
	if rt.Redis != nil {
		_ = rt.Redis.Close()
	}
	if rt.Pool != nil {
		rt.Pool.Close()
	}
}

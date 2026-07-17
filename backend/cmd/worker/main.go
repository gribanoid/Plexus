package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/plexus/backend/internal/config"
	"github.com/plexus/backend/internal/delivery"
	"github.com/plexus/backend/internal/logging"
	"github.com/plexus/backend/internal/metrics"
)

// plexus-worker runs asynq jobs (email, search index, webhooks) without serving the API.
// Use the same Docker image with command ["./plexus-worker"] and set RUN_WORKERS=false on API replicas.
func main() {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "error", err)
		os.Exit(1)
	}
	logging.Configure("plexus-worker", cfg.LogLevel)

	rt, err := delivery.NewRuntime(cfg)
	if err != nil {
		slog.Error("runtime", "error", err)
		os.Exit(1)
	}
	defer rt.Close()

	var metricsSrv *metrics.Server
	if cfg.MetricsEnabled {
		metricsSrv = metrics.NewServer(cfg.MetricsListenAddress, cfg.MetricsAuthToken)
		go func() {
			if err := metricsSrv.Start(); err != nil {
				slog.Error("metrics server error", "error", err)
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("plexus worker starting")
		if err := rt.JobServer.Start(); err != nil {
			slog.Error("job server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("worker shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if metricsSrv != nil {
		_ = metricsSrv.Shutdown(ctx)
	}
	rt.JobServer.Shutdown()
}

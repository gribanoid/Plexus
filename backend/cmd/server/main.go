package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/plexus/backend/internal/api"
	"github.com/plexus/backend/internal/config"
	"github.com/plexus/backend/internal/delivery"
	"github.com/plexus/backend/internal/logging"
	"github.com/plexus/backend/internal/metrics"
)

func main() {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "error", err)
		os.Exit(1)
	}
	logging.Configure("plexus-server", cfg.LogLevel)

	rt, err := delivery.NewRuntime(cfg)
	if err != nil {
		slog.Error("runtime", "error", err)
		os.Exit(1)
	}
	defer rt.Close()

	go rt.Hub.Run()

	if cfg.RunWorkers {
		go func() {
			if err := rt.JobServer.Start(); err != nil {
				slog.Error("job server error", "error", err)
			}
		}()
	}

	var metricsSrv *metrics.Server
	if cfg.MetricsEnabled {
		metricsSrv = metrics.NewServer(cfg.MetricsListenAddress, cfg.MetricsAuthToken)
		go func() {
			if err := metricsSrv.Start(); err != nil {
				slog.Error("metrics server error", "error", err)
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
		slog.Info("plexus server listening", "addr", addr, "workers", cfg.RunWorkers, "metrics", cfg.MetricsEnabled)
		if err := app.Listen(addr); err != nil {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if metricsSrv != nil {
		_ = metricsSrv.Shutdown(ctx)
	}
	if cfg.RunWorkers {
		rt.JobServer.Shutdown()
	}
	if err := app.Shutdown(); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}

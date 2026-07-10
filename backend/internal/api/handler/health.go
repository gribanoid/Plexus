package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthDeps struct {
	Pool  *pgxpool.Pool
	Redis *redis.Client
}

func DetailedHealth(deps HealthDeps) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()

		dbStatus := "ok"
		if err := deps.Pool.Ping(ctx); err != nil {
			dbStatus = "down"
		}

		redisStatus := "ok"
		if deps.Redis != nil {
			if err := deps.Redis.Ping(ctx).Err(); err != nil {
				redisStatus = "down"
			}
		} else {
			redisStatus = "not_configured"
		}

		overall := "ok"
		code := fiber.StatusOK
		if dbStatus != "ok" || redisStatus == "down" {
			overall = "degraded"
			code = fiber.StatusServiceUnavailable
		}

		return c.Status(code).JSON(fiber.Map{
			"status":  overall,
			"service": "plexus",
			"checks": fiber.Map{
				"database": dbStatus,
				"redis":    redisStatus,
			},
		})
	}
}

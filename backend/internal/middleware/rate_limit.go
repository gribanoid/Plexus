package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// RateLimitByIP limits requests per client IP using a Redis counter.
func RateLimitByIP(rdb *redis.Client, prefix string, limit int, window time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if rdb == nil {
			return c.Next()
		}

		key := fmt.Sprintf("ratelimit:%s:%s", prefix, c.IP())
		ctx := context.Background()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			return c.Next()
		}
		if count == 1 {
			_ = rdb.Expire(ctx, key, window).Err()
		}
		if count > int64(limit) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "too many requests, please try again later",
			})
		}

		return c.Next()
	}
}

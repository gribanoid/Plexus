package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/plexus/backend/internal/metrics"
)

// PrometheusHTTP records RED metrics for each request using the Fiber route path when available.
func PrometheusHTTP() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		status := c.Response().StatusCode()
		if err != nil {
			if e, ok := err.(*fiber.Error); ok {
				status = e.Code
			} else if status < 400 {
				status = fiber.StatusInternalServerError
			}
		}
		path := c.Route().Path
		if path == "" {
			path = c.Path()
		}
		metrics.ObserveHTTP(c.Method(), path, status, time.Since(start).Seconds())
		return err
	}
}

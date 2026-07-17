package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// RequireMetricsAuth protects scrape/ops endpoints with a shared bearer token.
// When token is empty (development only), access is allowed without auth.
func RequireMetricsAuth(token string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if token == "" {
			return c.Next()
		}
		auth := c.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") && strings.TrimPrefix(auth, "Bearer ") == token {
			return c.Next()
		}
		if c.Get("X-Metrics-Token") == token {
			return c.Next()
		}
		return fiber.ErrUnauthorized
	}
}

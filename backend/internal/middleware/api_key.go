package middleware

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/plexus/backend/internal/repository"
)

const ContextKeyAPIKeyID = "api_key_id"

// AuthOrAPIKey validates Bearer JWT or optional X-API-Key header.
func AuthOrAPIKey(jwtSecret string, repo *repository.Repo) fiber.Handler {
	jwtAuth := Auth(jwtSecret)

	return func(c *fiber.Ctx) error {
		if header := c.Get("Authorization"); header != "" {
			return jwtAuth(c)
		}

		rawKey := strings.TrimSpace(c.Get("X-API-Key"))
		if rawKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		keyHash := fmt.Sprintf("%x", sha256.Sum256([]byte(rawKey)))
		key, err := repo.GetAPIKeyByHash(c.Context(), keyHash)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "invalid api key",
				})
			}
			return err
		}

		if key.RevokedAt != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "api key revoked",
			})
		}
		if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "api key expired",
			})
		}

		c.Locals(ContextKeyUserID, key.CreatedBy)
		c.Locals(ContextKeyOrgID, key.OrgID)
		c.Locals(ContextKeyAPIKeyID, key.ID)
		_ = repo.TouchAPIKeyLastUsed(c.Context(), key.ID)

		return c.Next()
	}
}

// GetAPIKeyID returns the API key ID when the request was authenticated via X-API-Key.
func GetAPIKeyID(c *fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals(ContextKeyAPIKeyID).(uuid.UUID)
	return id, ok
}

package middleware

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/plexus/backend/internal/repository"
)

const ContextKeyAPIKeyOrgID = "api_key_org_id"
const ContextKeyAPIKeyScopes = "api_key_scopes"

// RequireOrgMember verifies the user belongs to the org in :orgSlug and sets org context.
// When authenticated via API key, the key must be bound to the same organization.
func RequireOrgMember(repo *repository.Repo) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := GetUserID(c)
		if !ok {
			return fiber.ErrUnauthorized
		}

		orgSlug := c.Params("orgSlug")
		org, err := repo.GetOrg(c.Context(), userID, orgSlug)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fiber.NewError(fiber.StatusNotFound, "organization not found")
			}
			return err
		}

		if keyOrg, ok := c.Locals(ContextKeyAPIKeyOrgID).(uuid.UUID); ok {
			if keyOrg != org.ID {
				return fiber.NewError(fiber.StatusForbidden, "api key is not valid for this organization")
			}
		}
		if err := enforceAPIKeyScopes(c); err != nil {
			return err
		}

		c.Locals(ContextKeyOrgID, org.ID)
		c.Locals(ContextKeyOrgRole, org.MyRole)
		return c.Next()
	}
}

func enforceAPIKeyScopes(c *fiber.Ctx) error {
	raw, ok := c.Locals(ContextKeyAPIKeyScopes).([]byte)
	if !ok || len(raw) == 0 {
		return nil
	}
	var scopes []string
	if err := json.Unmarshal(raw, &scopes); err != nil || len(scopes) == 0 {
		return nil
	}
	for _, s := range scopes {
		if s == "*" || strings.EqualFold(s, "admin") || strings.EqualFold(s, "full") {
			return nil
		}
	}
	method := c.Method()
	needWrite := method != fiber.MethodGet && method != fiber.MethodHead && method != fiber.MethodOptions
	for _, s := range scopes {
		if needWrite && (strings.EqualFold(s, "write") || strings.EqualFold(s, "readwrite")) {
			return nil
		}
		if !needWrite && (strings.EqualFold(s, "read") || strings.EqualFold(s, "write") || strings.EqualFold(s, "readwrite")) {
			return nil
		}
	}
	return fiber.NewError(fiber.StatusForbidden, "api key scope does not allow this operation")
}

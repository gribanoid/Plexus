package middleware

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/plexus/backend/internal/repository"
)

// RequireOrgMember verifies the user belongs to the org in :orgSlug and sets org context.
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

		c.Locals(ContextKeyOrgID, org.ID)
		c.Locals(ContextKeyOrgRole, org.MyRole)
		return c.Next()
	}
}

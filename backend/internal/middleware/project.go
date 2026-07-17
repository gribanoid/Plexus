package middleware

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/plexus/backend/internal/service/authz"
)

const (
	ContextKeyProjectID   = "project_id"
	ContextKeyProjectRole = "project_role"
)

// RequireProjectRead ensures the user can view the project in :orgSlug/:projectKey.
func RequireProjectRead(svc *authz.Service) fiber.Handler {
	return requireProject(svc, func(a *authz.ProjectAccess) bool { return a.CanRead })
}

// RequireProjectWrite ensures the user can mutate project data.
func RequireProjectWrite(svc *authz.Service) fiber.Handler {
	return requireProject(svc, func(a *authz.ProjectAccess) bool { return a.CanWrite })
}

// RequireProjectAdmin ensures project admin (or org admin).
func RequireProjectAdmin(svc *authz.Service) fiber.Handler {
	return requireProject(svc, func(a *authz.ProjectAccess) bool { return a.CanAdmin })
}

func requireProject(svc *authz.Service, ok func(*authz.ProjectAccess) bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, hasUser := GetUserID(c)
		if !hasUser {
			return fiber.ErrUnauthorized
		}
		access, err := svc.ResolveProjectAccess(c.Context(), userID, c.Params("orgSlug"), c.Params("projectKey"))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fiber.NewError(fiber.StatusNotFound, "project not found")
			}
			return err
		}
		if !ok(access) {
			return fiber.NewError(fiber.StatusForbidden, "insufficient project permissions")
		}
		c.Locals(ContextKeyProjectID, access.ProjectID)
		c.Locals(ContextKeyOrgID, access.OrgID)
		c.Locals(ContextKeyProjectRole, access.ProjectRole)
		return c.Next()
	}
}

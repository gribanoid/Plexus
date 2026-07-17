package handler

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/plexus/backend/internal/middleware"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,38}[a-z0-9]$`)

func (h *Handler) CreateOrg(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var body struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if body.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	autoSlug := body.Slug == ""
	if autoSlug {
		body.Slug = slugify(body.Name)
	}
	if !slugRegex.MatchString(body.Slug) {
		return fiber.NewError(fiber.StatusBadRequest, "slug must be 3–40 lowercase alphanumeric characters or hyphens")
	}
	if autoSlug {
		unique, err := h.Repo.UniqueOrgSlug(c.Context(), body.Slug)
		if err != nil {
			return err
		}
		body.Slug = unique
	} else {
		exists, err := h.Repo.OrgSlugExists(c.Context(), body.Slug)
		if err != nil {
			return err
		}
		if exists {
			return fiber.NewError(fiber.StatusConflict, "slug already taken")
		}
	}

	orgID := uuid.New()
	if err := h.Repo.CreateOrg(c.Context(), orgID, userID, body.Slug, body.Name); err != nil {
		return fiber.NewError(fiber.StatusConflict, "slug already taken")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":   orgID,
		"slug": body.Slug,
		"name": body.Name,
		"plan": "free",
	})
}

func (h *Handler) ListMyOrgs(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	orgs, err := h.Repo.ListMyOrgs(c.Context(), userID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"items": mapSlice(orgs, toOrgDTO)})
}

func (h *Handler) GetOrg(c *fiber.Ctx) error {
	orgSlug := c.Params("orgSlug")
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	org, err := h.Repo.GetOrg(c.Context(), userID, orgSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "organization not found")
		}
		return err
	}
	return c.JSON(toOrgDetailDTO(*org))
}

func (h *Handler) UpdateOrg(c *fiber.Ctx) error {
	orgSlug := c.Params("orgSlug")
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	if admin, err := h.isOrgAdmin(c.Context(), orgSlug, userID); err != nil {
		return err
	} else if !admin {
		return fiber.NewError(fiber.StatusForbidden, "insufficient permissions")
	}

	var body struct {
		Name    *string `json:"name"`
		LogoURL *string `json:"logo_url"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	return h.Repo.UpdateOrg(c.Context(), orgSlug, body.Name, body.LogoURL)
}

func (h *Handler) ListOrgMembers(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	members, err := h.Repo.ListOrgMembers(c.Context(), userID, c.Params("orgSlug"))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"items": mapSlice(members, toOrgMemberDTO)})
}

func (h *Handler) InviteMember(c *fiber.Ctx) error {
	orgSlug := c.Params("orgSlug")
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	if admin, err := h.isOrgAdmin(c.Context(), orgSlug, userID); err != nil {
		return err
	} else if !admin {
		return fiber.NewError(fiber.StatusForbidden, "insufficient permissions")
	}

	var body struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := c.BodyParser(&body); err != nil || body.Email == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email is required")
	}
	if body.Role == "" {
		body.Role = "member"
	}
	switch body.Role {
	case "admin", "member", "guest":
		// ok — owner can only be granted by an existing owner
	case "owner":
		callerRole, _ := c.Locals(middleware.ContextKeyOrgRole).(string)
		if callerRole != "owner" {
			return fiber.NewError(fiber.StatusForbidden, "only owners can assign the owner role")
		}
	default:
		return fiber.NewError(fiber.StatusBadRequest, "role must be owner, admin, member, or guest")
	}

	inviteeID, err := h.Repo.GetUserIDByEmail(c.Context(), body.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}
		return err
	}

	orgID, err := h.Repo.GetOrgIDBySlug(c.Context(), orgSlug)
	if err != nil {
		return err
	}

	if err := h.Repo.UpsertOrgMember(c.Context(), orgID, inviteeID, body.Role); err != nil {
		return err
	}

	h.logAudit(c.Context(), &orgID, userID, "invite", "org_member", &inviteeID, map[string]any{
		"email": body.Email,
		"role":  body.Role,
	})

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) RemoveMember(c *fiber.Ctx) error {
	orgSlug := c.Params("orgSlug")
	targetUserID, err := uuid.Parse(c.Params("userID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid user id")
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	if admin, err := h.isOrgAdmin(c.Context(), orgSlug, userID); err != nil {
		return err
	} else if !admin {
		return fiber.NewError(fiber.StatusForbidden, "insufficient permissions")
	}

	if err := h.Repo.RemoveOrgMember(c.Context(), orgSlug, targetUserID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) isOrgAdmin(ctx context.Context, orgSlug string, userID uuid.UUID) (bool, error) {
	return h.Repo.IsOrgAdmin(ctx, orgSlug, userID)
}

func slugify(s string) string {
	s = strings.ToLower(s)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}

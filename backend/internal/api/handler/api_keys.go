package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/plexus/backend/internal/middleware"
)

func (h *Handler) CreateAPIKey(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	if admin, err := h.isOrgAdmin(c.Context(), c.Params("orgSlug"), userID); err != nil {
		return err
	} else if !admin {
		return fiber.NewError(fiber.StatusForbidden, "only org admins can create API keys")
	}

	var body struct {
		Name      string     `json:"name"`
		Scopes    []string   `json:"scopes"`
		ExpiresAt *time.Time `json:"expires_at"`
	}
	if err := c.BodyParser(&body); err != nil || body.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}

	org, err := h.Repo.GetOrg(c.Context(), userID, c.Params("orgSlug"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "organization not found")
		}
		return err
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return err
	}
	rawKey := "plx_" + hex.EncodeToString(raw)
	prefix := rawKey[:12]
	keyHash := fmt.Sprintf("%x", sha256.Sum256([]byte(rawKey)))

	id := uuid.New()
	if err := h.Repo.CreateAPIKey(c.Context(), id, org.ID, userID, body.Name, keyHash, prefix, body.Scopes, body.ExpiresAt); err != nil {
		return err
	}

	h.logAudit(c.Context(), &org.ID, userID, "create", "api_key", &id, map[string]any{
		"name":   body.Name,
		"prefix": prefix,
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":     id,
		"name":   body.Name,
		"prefix": prefix,
		"key":    rawKey,
		"scopes": body.Scopes,
	})
}

func (h *Handler) ListAPIKeys(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	if admin, err := h.isOrgAdmin(c.Context(), c.Params("orgSlug"), userID); err != nil {
		return err
	} else if !admin {
		return fiber.NewError(fiber.StatusForbidden, "only org admins can list API keys")
	}

	org, err := h.Repo.GetOrg(c.Context(), userID, c.Params("orgSlug"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "organization not found")
		}
		return err
	}

	items, err := h.Repo.ListAPIKeys(c.Context(), org.ID)
	if err != nil {
		return err
	}
	out := make([]fiber.Map, 0, len(items))
	for _, k := range items {
		var scopes any
		if len(k.Scopes) > 0 {
			_ = json.Unmarshal(k.Scopes, &scopes)
		}
		m := fiber.Map{
			"id":         k.ID,
			"name":       k.Name,
			"prefix":     k.Prefix,
			"scopes":     scopes,
			"created_by": k.CreatedBy,
			"created_at": k.CreatedAt,
		}
		if k.LastUsedAt != nil {
			m["last_used_at"] = *k.LastUsedAt
		}
		if k.ExpiresAt != nil {
			m["expires_at"] = *k.ExpiresAt
		}
		if k.RevokedAt != nil {
			m["revoked_at"] = *k.RevokedAt
		}
		out = append(out, m)
	}
	return c.JSON(fiber.Map{"items": out})
}

func (h *Handler) RevokeAPIKey(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}
	if admin, err := h.isOrgAdmin(c.Context(), c.Params("orgSlug"), userID); err != nil {
		return err
	} else if !admin {
		return fiber.NewError(fiber.StatusForbidden, "only org admins can revoke API keys")
	}

	keyID, err := uuid.Parse(c.Params("keyID"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid key id")
	}

	org, err := h.Repo.GetOrg(c.Context(), userID, c.Params("orgSlug"))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "organization not found")
		}
		return err
	}

	if err := h.Repo.RevokeAPIKey(c.Context(), keyID, org.ID); err != nil {
		return err
	}

	h.logAudit(c.Context(), &org.ID, userID, "revoke", "api_key", &keyID, nil)
	return c.SendStatus(fiber.StatusNoContent)
}

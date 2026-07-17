package handler

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/plexus/backend/internal/middleware"
	"github.com/plexus/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var req registerRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email, password and display_name are required")
	}
	if len(req.Password) < 8 {
		return fiber.NewError(fiber.StatusBadRequest, "password must be at least 8 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	userID := uuid.New()
	orgSlug, err := h.Repo.UniqueOrgSlug(c.Context(), slugify(strings.Split(req.Email, "@")[0]))
	if err != nil {
		return err
	}

	err = h.Repo.RegisterUser(c.Context(), repository.RegisterInput{
		UserID:       userID,
		Email:        req.Email,
		PasswordHash: string(hash),
		DisplayName:  req.DisplayName,
		OrgID:        uuid.New(),
		OrgSlug:      orgSlug,
		OrgName:      req.DisplayName + "'s Workspace",
		ProjectID:    uuid.New(),
	})
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, "email already in use")
	}

	tokens, err := h.issueTokens(c.Context(), userID, req.Email)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(tokens)
}

func resolveLoginEmail(email string) string {
	trimmed := strings.TrimSpace(email)
	if strings.EqualFold(trimmed, "admin") {
		return "admin@plexus.local"
	}
	return trimmed
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	email := resolveLoginEmail(req.Email)
	creds, err := h.Repo.GetUserCredentials(c.Context(), email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
		}
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(req.Password)); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
	}

	tokens, err := h.issueTokens(c.Context(), creds.ID, creds.Email)
	if err != nil {
		return err
	}
	return c.JSON(tokens)
}

func (h *Handler) RefreshToken(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&body); err != nil || body.RefreshToken == "" {
		return fiber.NewError(fiber.StatusBadRequest, "refresh_token is required")
	}

	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(body.RefreshToken)))

	userID, email, err := h.Repo.ConsumeRefreshToken(c.Context(), tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired refresh token")
		}
		return err
	}

	tokens, err := h.issueTokens(c.Context(), userID, email)
	if err != nil {
		return err
	}
	return c.JSON(tokens)
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&body); err == nil && body.RefreshToken != "" {
		tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(body.RefreshToken)))
		_ = h.Repo.DeleteRefreshToken(c.Context(), tokenHash)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) GetMe(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	u, err := h.Repo.GetUserProfile(c.Context(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}
		return err
	}
	return c.JSON(toUserDTO(*u))
}

func (h *Handler) UpdateMe(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return fiber.ErrUnauthorized
	}

	var body struct {
		DisplayName *string `json:"display_name"`
		AvatarURL   *string `json:"avatar_url"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if err := h.Repo.UpdateUserProfile(c.Context(), userID, body.DisplayName, body.AvatarURL); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) issueTokens(ctx context.Context, userID uuid.UUID, email string) (*tokenPair, error) {
	const accessTTL = 15 * time.Minute
	const refreshTTL = 30 * 24 * time.Hour

	claims := middleware.JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(h.JWTSecret))
	if err != nil {
		return nil, err
	}

	refreshRaw := uuid.New().String()
	refreshHash := fmt.Sprintf("%x", sha256.Sum256([]byte(refreshRaw)))

	if err := h.Repo.InsertRefreshToken(ctx, uuid.New(), userID, refreshHash, time.Now().Add(refreshTTL)); err != nil {
		return nil, err
	}

	return &tokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshRaw,
		ExpiresIn:    int(accessTTL.Seconds()),
	}, nil
}

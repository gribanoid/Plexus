package middleware_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/plexus/backend/internal/middleware"
)

func TestRequireOrgRole_AllowsMatchingRole(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(middleware.ContextKeyOrgRole, "admin")
		return c.Next()
	})
	app.Get("/", middleware.RequireOrgRole("owner", "admin"), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRequireOrgRole_DeniesGuest(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(middleware.ContextKeyOrgRole, "guest")
		return c.Next()
	})
	app.Post("/", middleware.RequireOrgRole("owner", "admin", "member"), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("POST", "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

func TestParseJWTToken_ValidToken(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New()
	token := signTestToken(t, userID, secret)

	parsed, err := middleware.ParseJWTToken(token, secret)
	if err != nil {
		t.Fatal(err)
	}
	if parsed != userID {
		t.Fatalf("parsed user mismatch: %s vs %s", parsed, userID)
	}
}

func TestAuth_RejectsMissingHeader(t *testing.T) {
	app := fiber.New()
	app.Get("/", middleware.Auth("secret"), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func signTestToken(t *testing.T, userID uuid.UUID, secret string) string {
	t.Helper()
	claims := middleware.JWTClaims{
		UserID: userID,
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			Subject:   userID.String(),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

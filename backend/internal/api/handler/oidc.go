package handler

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/plexus/backend/internal/auth"
	"github.com/plexus/backend/internal/repository"
)

const oidcStateCookie = "oidc_state"

func (h *Handler) OIDCLogin(c *fiber.Ctx) error {
	if h.OIDC == nil || !h.OIDC.Configured() {
		return fiber.NewError(fiber.StatusNotImplemented, "OIDC is not configured")
	}

	state, err := auth.NewState()
	if err != nil {
		return err
	}

	authURL, err := h.OIDC.AuthorizationURL(state)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "failed to discover OIDC provider: "+err.Error())
	}

	c.Cookie(&fiber.Cookie{
		Name:     oidcStateCookie,
		Value:    state,
		HTTPOnly: true,
		Secure:   c.Protocol() == "https",
		SameSite: "Lax",
		MaxAge:   600,
		Path:     "/",
	})

	return c.Redirect(authURL, fiber.StatusFound)
}

func (h *Handler) OIDCCallback(c *fiber.Ctx) error {
	if h.OIDC == nil || !h.OIDC.Configured() {
		return fiber.NewError(fiber.StatusNotImplemented, "OIDC is not configured")
	}

	if errMsg := c.Query("error"); errMsg != "" {
		return fiber.NewError(fiber.StatusUnauthorized, errMsg)
	}

	code := c.Query("code")
	if code == "" {
		return fiber.NewError(fiber.StatusBadRequest, "authorization code is required")
	}

	state := c.Query("state")
	cookieState := c.Cookies(oidcStateCookie)
	if state == "" || cookieState == "" || state != cookieState {
		return fiber.NewError(fiber.StatusBadRequest, "invalid OAuth state")
	}
	c.ClearCookie(oidcStateCookie)

	tokens, err := h.OIDC.ExchangeCode(c.Context(), code)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "token exchange failed: "+err.Error())
	}

	accessToken, _ := tokens["access_token"].(string)
	var claims map[string]any
	if accessToken != "" {
		claims, err = h.OIDC.UserInfo(c.Context(), accessToken)
		if err != nil {
			return fiber.NewError(fiber.StatusBadGateway, "userinfo request failed: "+err.Error())
		}
	}

	email, _ := claims["email"].(string)
	if email == "" {
		return fiber.NewError(fiber.StatusBadRequest, "email claim is required")
	}

	creds, err := h.Repo.GetUserCredentials(c.Context(), email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			creds, err = h.provisionOIDCUser(c, email, claims)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	tokenPair, err := h.issueTokens(c.Context(), creds.ID, creds.Email)
	if err != nil {
		return err
	}

	redirectURL := h.OIDC.CallbackRedirectURL(tokenPair.AccessToken, tokenPair.RefreshToken)
	if h.FrontendURL != "" {
		q := url.Values{}
		q.Set("access_token", tokenPair.AccessToken)
		q.Set("refresh_token", tokenPair.RefreshToken)
		redirectURL = strings.TrimRight(h.FrontendURL, "/") + "/auth/callback?" + q.Encode()
	}

	return c.Redirect(redirectURL, fiber.StatusFound)
}

func (h *Handler) provisionOIDCUser(c *fiber.Ctx, email string, claims map[string]any) (*repository.UserCredentials, error) {
	displayName := auth.DisplayNameFromClaims(claims, email)
	passwordHash, err := randomUnusablePasswordHash()
	if err != nil {
		return nil, err
	}

	userID := uuid.New()
	orgSlug, err := h.Repo.UniqueOrgSlug(c.Context(), slugify(strings.Split(email, "@")[0]))
	if err != nil {
		return nil, err
	}

	err = h.Repo.RegisterUser(c.Context(), repository.RegisterInput{
		UserID:       userID,
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
		OrgID:        uuid.New(),
		OrgSlug:      orgSlug,
		OrgName:      displayName + "'s Workspace",
		ProjectID:    uuid.New(),
	})
	if err != nil {
		return nil, fiber.NewError(fiber.StatusConflict, "could not provision user: "+err.Error())
	}

	return &repository.UserCredentials{ID: userID, Email: email, PasswordHash: passwordHash}, nil
}

func randomUnusablePasswordHash() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "$oidc$" + base64.RawURLEncoding.EncodeToString(b), nil
}

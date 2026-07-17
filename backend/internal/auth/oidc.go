package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// OIDCConfig holds OpenID Connect provider settings loaded from environment.
type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	FrontendURL  string
	Scopes       []string
}

// Provider caches OIDC discovery document endpoints.
type Provider struct {
	Config       OIDCConfig
	authorizeURL string
	tokenURL     string
	userInfoURL  string
	mu           sync.RWMutex
	discovered   bool
}

// LoadOIDCFromEnv reads OIDC_ISSUER, OIDC_CLIENT_ID, OIDC_CLIENT_SECRET and optional OIDC_REDIRECT_URI.
func LoadOIDCFromEnv() *Provider {
	issuer := strings.TrimSpace(os.Getenv("OIDC_ISSUER"))
	clientID := strings.TrimSpace(os.Getenv("OIDC_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("OIDC_CLIENT_SECRET"))
	redirectURI := strings.TrimSpace(os.Getenv("OIDC_REDIRECT_URI"))
	if redirectURI == "" {
		redirectURI = "http://localhost:8080/api/v1/auth/oidc/callback"
	}
	frontendURL := strings.TrimSpace(os.Getenv("FRONTEND_URL"))
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	return &Provider{
		Config: OIDCConfig{
			Issuer:       issuer,
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURI:  redirectURI,
			FrontendURL:  strings.TrimRight(frontendURL, "/"),
			Scopes:       []string{"openid", "profile", "email"},
		},
	}
}

// Configured reports whether all required OIDC settings are present.
func (p *Provider) Configured() bool {
	if p == nil {
		return false
	}
	return p.Config.Issuer != "" && p.Config.ClientID != "" && p.Config.ClientSecret != ""
}

// AuthorizationURL builds the provider authorize URL with a random state nonce.
func (p *Provider) AuthorizationURL(state string) (string, error) {
	if err := p.discover(context.Background()); err != nil {
		return "", err
	}

	p.mu.RLock()
	authURL := p.authorizeURL
	p.mu.RUnlock()

	q := url.Values{}
	q.Set("client_id", p.Config.ClientID)
	q.Set("response_type", "code")
	q.Set("scope", strings.Join(p.Config.Scopes, " "))
	q.Set("redirect_uri", p.Config.RedirectURI)
	q.Set("state", state)
	return authURL + "?" + q.Encode(), nil
}

// ExchangeCode trades an authorization code for tokens at the provider token endpoint.
func (p *Provider) ExchangeCode(ctx context.Context, code string) (map[string]any, error) {
	if err := p.discover(ctx); err != nil {
		return nil, err
	}

	p.mu.RLock()
	tokenURL := p.tokenURL
	p.mu.RUnlock()

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", p.Config.RedirectURI)
	form.Set("client_id", p.Config.ClientID)
	form.Set("client_secret", p.Config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var tokens map[string]any
	if err := json.Unmarshal(body, &tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}

// UserInfo fetches claims for the given access token.
func (p *Provider) UserInfo(ctx context.Context, accessToken string) (map[string]any, error) {
	if err := p.discover(ctx); err != nil {
		return nil, err
	}

	p.mu.RLock()
	userInfoURL := p.userInfoURL
	p.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("userinfo endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var claims map[string]any
	if err := json.Unmarshal(body, &claims); err != nil {
		return nil, err
	}
	return claims, nil
}

// DisplayNameFromClaims extracts a human-readable name from OIDC userinfo claims.
func DisplayNameFromClaims(claims map[string]any, email string) string {
	for _, key := range []string{"name", "preferred_username", "given_name"} {
		if v, ok := claims[key].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	if email != "" {
		if at := strings.Index(email, "@"); at > 0 {
			return email[:at]
		}
		return email
	}
	return "User"
}

// CallbackRedirectURL returns the frontend OAuth callback path (without tokens).
func (p *Provider) CallbackRedirectURL() string {
	base := strings.TrimRight(p.Config.FrontendURL, "/")
	if base == "" {
		base = "http://localhost:3000"
	}
	return base + "/auth/callback"
}

// NewState generates a cryptographically random OAuth state value.
func NewState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

type discoveryDocument struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserInfoEndpoint      string `json:"userinfo_endpoint"`
}

func (p *Provider) discover(ctx context.Context) error {
	p.mu.RLock()
	if p.discovered {
		p.mu.RUnlock()
		return nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.discovered {
		return nil
	}

	wellKnown := strings.TrimRight(p.Config.Issuer, "/") + "/.well-known/openid-configuration"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wellKnown, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("discovery document returned %d", resp.StatusCode)
	}

	var doc discoveryDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return err
	}
	if doc.AuthorizationEndpoint == "" || doc.TokenEndpoint == "" {
		return fmt.Errorf("incomplete OIDC discovery document")
	}

	p.authorizeURL = doc.AuthorizationEndpoint
	p.tokenURL = doc.TokenEndpoint
	p.userInfoURL = doc.UserInfoEndpoint
	p.discovered = true
	return nil
}

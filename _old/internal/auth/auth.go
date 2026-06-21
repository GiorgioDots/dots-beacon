// Package auth provides shared Keycloak (OIDC) token authentication for the web
// APIs. Keycloak handles login; the APIs are resource servers that only verify
// the bearer access token on each request. All services share this package, so
// authentication is wired the same everywhere — set a couple of env vars and add
// the middleware.
package auth

import (
	"context"
	"fmt"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
)

// Config controls how incoming tokens are verified.
type Config struct {
	// IssuerURL is the Keycloak realm issuer, e.g.
	// http://localhost:8080/realms/dots-beacon. OIDC discovery and JWKS are
	// resolved from it.
	IssuerURL string
	// ClientID is the expected audience (aud) of the token. If empty, the
	// audience check is skipped (handy for Keycloak access tokens, whose aud is
	// often "account"); the token signature, issuer, and expiry are still
	// verified.
	ClientID string
	// SkipIssuerCheck disables the issuer match. Only needed when Keycloak's
	// external URL differs from the one tokens are minted with (e.g. container
	// networking). Prefer aligning the issuer URL instead.
	SkipIssuerCheck bool
}

// ConfigFromEnv reads Config from KEYCLOAK_ISSUER_URL / KEYCLOAK_CLIENT_ID.
func ConfigFromEnv() Config {
	return Config{
		IssuerURL: os.Getenv("KEYCLOAK_ISSUER_URL"),
		ClientID:  os.Getenv("KEYCLOAK_CLIENT_ID"),
	}
}

// Enabled reports whether enough config is present to set up authentication.
func (c Config) Enabled() bool { return c.IssuerURL != "" }

// Authenticator verifies Keycloak-issued JWTs. It is safe for concurrent use;
// the underlying provider caches and rotates the JWKS automatically.
type Authenticator struct {
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
}

// New performs OIDC discovery against the issuer and builds a token verifier.
// It needs Keycloak to be reachable at call time.
func New(ctx context.Context, cfg Config) (*Authenticator, error) {
	if cfg.IssuerURL == "" {
		return nil, fmt.Errorf("auth: IssuerURL is required")
	}

	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("auth: discover oidc provider: %w", err)
	}

	oidcCfg := &oidc.Config{
		ClientID:          cfg.ClientID,
		SkipClientIDCheck: cfg.ClientID == "",
		SkipIssuerCheck:   cfg.SkipIssuerCheck,
	}

	return &Authenticator{
		provider: provider,
		verifier: provider.Verifier(oidcCfg),
	}, nil
}

// Verify checks a raw JWT and returns the authenticated user.
func (a *Authenticator) Verify(ctx context.Context, rawToken string) (User, error) {
	token, err := a.verifier.Verify(ctx, rawToken)
	if err != nil {
		return User{}, fmt.Errorf("auth: verify token: %w", err)
	}

	var cl claims
	if err := token.Claims(&cl); err != nil {
		return User{}, fmt.Errorf("auth: parse claims: %w", err)
	}
	return cl.toUser(), nil
}

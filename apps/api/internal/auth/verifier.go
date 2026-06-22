package auth

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/giorgiodots/dots-beacon/api/internal/config"
)

type AuthVerifier struct {
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
}

type Claims struct {
	Sub string `json:"sub"`
}

func NewAuthVerifier(ctx context.Context, cfg config.Config) (*AuthVerifier, error) {
	provider, err := oidc.NewProvider(ctx, cfg.OidUrl)
	if err != nil {
		return nil, err
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.OidClientId})

	auth := &AuthVerifier{
		provider: provider,
		verifier: verifier,
	}

	return auth, nil
}

func (v *AuthVerifier) Verify(ctx context.Context, token string) (*Claims, error) {
	idToken, err := v.verifier.Verify(ctx, token)
	if err != nil {
		return nil, err
	}

	var claims Claims
	err = idToken.Claims(&claims)
	if err != nil {
		return nil, err
	}
	return &claims, nil
}

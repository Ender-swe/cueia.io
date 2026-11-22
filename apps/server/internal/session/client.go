package session

import (
	"context"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

func Clerk(ctx context.Context, issuer, party string) VerifyFunc {
	issuer = strings.TrimSuffix(issuer, "/")
	keys := oidc.NewRemoteKeySet(ctx, issuer+"/.well-known/jwks.json")
	verifier := oidc.NewVerifier(issuer, keys, &oidc.Config{SkipClientIDCheck: true, SupportedSigningAlgs: []string{"RS256"}})
	return func(ctx context.Context, raw string) (string, error) {
		token, err := verifier.Verify(ctx, raw)
		if err != nil {
			return "", ErrUnauthorized
		}
		var claims struct {
			Party   string `json:"azp"`
			Session string `json:"sid"`
		}
		if err = token.Claims(&claims); err != nil || claims.Party != party || claims.Session == "" || token.Subject == "" {
			return "", ErrUnauthorized
		}
		return token.Subject, nil
	}
}

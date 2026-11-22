package session

import (
	"context"
	"errors"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUnauthorized = errors.New("unauthorized")

// VerifyFunc returns the Clerk subject for a verified session token.
type VerifyFunc func(ctx context.Context, rawToken string) (subject string, err error)

// Verify checks the Clerk token and returns its subject.
func (s *Service) Verify(ctx context.Context, clerkToken string) (string, error) {
	if clerkToken == "" {
		return "", ErrUnauthorized
	}
	sub, err := s.verify(ctx, clerkToken)
	if err != nil {
		return "", ErrUnauthorized
	}
	return sub, nil
}

// NewService creates a session service that shares Clerk signing keys across requests.
func NewService(ctx context.Context, db *pgxpool.Pool, cfg Config, provision func(context.Context, string) (string, error)) *Service {
	issuer := strings.TrimSuffix(cfg.ClerkIssuer, "/")
	keys := oidc.NewRemoteKeySet(ctx, issuer+"/.well-known/jwks.json")
	clerk := clerkVerifier{
		tokens: oidc.NewVerifier(issuer, keys, &oidc.Config{
			SkipClientIDCheck:    true,
			SupportedSigningAlgs: []string{"RS256"},
		}),
		authorizedParty: cfg.AuthorizedParty,
	}
	return &Service{db: db, verify: clerk.subject, provision: provision}
}

type clerkVerifier struct {
	tokens          *oidc.IDTokenVerifier
	authorizedParty string
}

func (v clerkVerifier) subject(ctx context.Context, raw string) (string, error) {
	token, err := v.tokens.Verify(ctx, raw)
	if err != nil {
		return "", err
	}
	var claims struct {
		Party   string `json:"azp"`
		Session string `json:"sid"`
	}
	if err := token.Claims(&claims); err != nil {
		return "", err
	}
	if claims.Party != v.authorizedParty || claims.Session == "" || token.Subject == "" {
		return "", ErrUnauthorized
	}
	return token.Subject, nil
}

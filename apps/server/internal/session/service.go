package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUnauthorized = errors.New("unauthorized")

type Service struct {
	memory          memory
	verifier        *oidc.IDTokenVerifier
	authorizedParty string
	provision       func(context.Context, string) (string, error)
}

// Verify checks the Clerk token and returns its subject.
func (s *Service) Verify(ctx context.Context, clerkToken string) (string, error) {
	if clerkToken == "" {
		return "", ErrUnauthorized
	}
	token, err := s.verifier.Verify(ctx, clerkToken)
	if err != nil {
		return "", ErrUnauthorized
	}
	var claims struct {
		Party   string `json:"azp"`
		Session string `json:"sid"`
	}
	if err := token.Claims(&claims); err != nil || claims.Party != s.authorizedParty || claims.Session == "" || token.Subject == "" {
		return "", ErrUnauthorized
	}
	return token.Subject, nil
}

func NewService(ctx context.Context, db *pgxpool.Pool, provision func(context.Context, string) (string, error)) (*Service, error) {
	cfg, missing := loadConfig()
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing settings: %s", strings.Join(missing, ", "))
	}
	issuer := strings.TrimSuffix(cfg.ClerkIssuer, "/")
	keys := oidc.NewRemoteKeySet(ctx, issuer+"/.well-known/jwks.json")
	verifier := oidc.NewVerifier(issuer, keys, &oidc.Config{
		SkipClientIDCheck:    true,
		SupportedSigningAlgs: []string{"RS256"},
	})
	return &Service{memory: memory{db: db}, verifier: verifier, authorizedParty: cfg.AuthorizedParty, provision: provision}, nil
}

func (s *Service) Exchange(ctx context.Context, clerkToken string) (Credential, error) {
	sub, err := s.Verify(ctx, clerkToken)
	if err != nil {
		return Credential{}, err
	}
	w, err := s.provision(ctx, sub)
	if err != nil {
		return Credential{}, err
	}
	var secret [tokenBytes]byte
	rand.Read(secret[:])
	token := base64.RawURLEncoding.EncodeToString(secret[:])
	hash := sha256.Sum256([]byte(token))
	expires, err := s.memory.createSession(ctx, hash[:], sub, w)
	if err != nil {
		return Credential{}, err
	}
	return Credential{token, expires}, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", ErrUnauthorized
	}
	hash := sha256.Sum256([]byte(token))
	return s.memory.workspaceForSession(ctx, hash[:])
}

func (s *Service) Revoke(ctx context.Context, token string) error {
	hash := sha256.Sum256([]byte(token))
	return s.memory.revokeSession(ctx, hash[:])
}

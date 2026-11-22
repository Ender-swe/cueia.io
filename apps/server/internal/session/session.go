package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db        *pgxpool.Pool
	verify    VerifyFunc
	provision func(context.Context, string) (string, error)
}

type Credential struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

var ErrUnauthorized = errors.New("unauthorized")

// VerifyFunc returns the Clerk subject for a verified session token.
type VerifyFunc func(ctx context.Context, rawToken string) (subject string, err error)

func NewService(db *pgxpool.Pool, verify VerifyFunc, provision func(context.Context, string) (string, error)) *Service {
	return &Service{db: db, verify: verify, provision: provision}
}

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

func (s *Service) Exchange(ctx context.Context, clerkToken string) (Credential, error) {
	sub, err := s.Verify(ctx, clerkToken)
	if err != nil {
		return Credential{}, err
	}
	w, err := s.provision(ctx, sub)
	if err != nil {
		return Credential{}, err
	}
	var secret [32]byte
	rand.Read(secret[:])
	token := base64.RawURLEncoding.EncodeToString(secret[:])
	hash := sha256.Sum256([]byte(token))
	var expires time.Time
	err = s.db.QueryRow(ctx, `INSERT INTO sessions(token_digest,user_id,expires_at)
 SELECT $1,id,now()+interval '12 hours' FROM users WHERE clerk_user_id=$2 AND workspace_id=$3 RETURNING expires_at`, hash[:], sub, w).Scan(&expires)
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
	var w string
	err := s.db.QueryRow(ctx, `SELECT u.workspace_id FROM sessions s JOIN users u ON u.id=s.user_id WHERE token_digest=$1 AND revoked_at IS NULL AND expires_at>now()`, hash[:]).Scan(&w)
	return w, err
}

func (s *Service) Revoke(ctx context.Context, token string) error {
	hash := sha256.Sum256([]byte(token))
	_, err := s.db.Exec(ctx, "UPDATE sessions SET revoked_at=now() WHERE token_digest=$1 AND revoked_at IS NULL", hash[:])
	return err
}


package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db        *pgxpool.Pool
	verify    VerifyFunc
	provision func(context.Context, string) (string, error)
}

func NewServiceWithVerifier(db *pgxpool.Pool, verify VerifyFunc, provision func(context.Context, string) (string, error)) *Service {
	return &Service{db: db, verify: verify, provision: provision}
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
	var expires time.Time
	err = s.db.QueryRow(ctx, `INSERT INTO sessions(token_digest,user_id,expires_at)
 SELECT $1,id,now()+$4::text::interval FROM users WHERE clerk_user_id=$2 AND workspace_id=$3 RETURNING expires_at`, hash[:], sub, w, sessionLifetime).Scan(&expires)
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

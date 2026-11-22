package session

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type memory struct {
	db *pgxpool.Pool
}

func (m memory) createSession(ctx context.Context, digest []byte, subject, workspaceID string) (time.Time, error) {
	var expires time.Time
	err := m.db.QueryRow(ctx, `
        INSERT INTO sessions(token_digest, user_id, expires_at)
        SELECT $1, id, now() + $4::text::interval FROM users
        WHERE clerk_user_id = $2 AND workspace_id = $3
        RETURNING expires_at`, digest, subject, workspaceID, sessionLifetime).Scan(&expires)
	return expires, err
}

func (m memory) workspaceForSession(ctx context.Context, digest []byte) (string, error) {
	var workspaceID string
	err := m.db.QueryRow(ctx, `
        SELECT u.workspace_id FROM sessions s JOIN users u ON u.id = s.user_id
        WHERE s.token_digest = $1 AND s.revoked_at IS NULL AND s.expires_at > now()`, digest).Scan(&workspaceID)
	return workspaceID, err
}

func (m memory) revokeSession(ctx context.Context, digest []byte) error {
	_, err := m.db.Exec(ctx, `
        UPDATE sessions SET revoked_at = now()
        WHERE token_digest = $1 AND revoked_at IS NULL`, digest)
	return err
}

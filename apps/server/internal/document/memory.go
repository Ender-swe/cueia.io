package document

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type memory struct {
	db *pgxpool.Pool
}

func (m *memory) loadSigningKey(ctx context.Context, candidate []byte) ([]byte, error) {
	var key []byte
	err := pgx.BeginFunc(ctx, m.db, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `
			INSERT INTO deployment_signing_keys(singleton, key_material)
			VALUES (true, $1) ON CONFLICT DO NOTHING`, candidate); err != nil {
			return err
		}
		return tx.QueryRow(ctx, "SELECT key_material FROM deployment_signing_keys WHERE singleton").Scan(&key)
	})
	return key, err
}

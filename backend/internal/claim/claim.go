package claim

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
)

func AtVersion(ctx context.Context, conn *pgx.Conn, id string, version int64) (bool, error) {
	tag, err := conn.Exec(ctx, "UPDATE claim_records SET status='claimed',version=version+1 WHERE id=$1 AND version=$2 AND status='eligible'", id, version)
	return tag.RowsAffected() == 1, err
}

func Next(ctx context.Context, conn *pgx.Conn) error {
	var id string
	var version int64
	err := conn.QueryRow(ctx, "SELECT id,version FROM claim_records WHERE status='eligible' ORDER BY id LIMIT 1").Scan(&id, &version)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = AtVersion(ctx, conn, id, version)
	return err
}

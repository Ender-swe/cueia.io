package database

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/Ender-swe/cueia.io/apps/server/internal/config"
	"github.com/jackc/pgx/v5"
)

func Open(ctx context.Context) (*pgx.Conn, error) {
	url := config.DatabaseURL()
	if url == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return nil, errors.New("cannot connect to PostgreSQL")
	}
	return conn, nil
}

type migration struct {
	version int64
	path    string
}

func Migrate(ctx context.Context, conn *pgx.Conn, dir fs.FS) error {
	paths, err := fs.Glob(dir, "*.sql")
	if err != nil || len(paths) == 0 {
		return errors.New("no SQL migrations found")
	}
	var migrations []migration
	for _, path := range paths {
		prefix, _, ok := strings.Cut(path, "_")
		version, err := strconv.ParseInt(prefix, 10, 64)
		if !ok || err != nil || version < 1 {
			return fmt.Errorf("invalid migration filename %s", path)
		}
		migrations = append(migrations, migration{version, path})
	}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].version < migrations[j].version })
	return pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(37151694)"); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations (version bigint PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())"); err != nil {
			return err
		}
		for i, m := range migrations {
			if i > 0 && m.version == migrations[i-1].version {
				return fmt.Errorf("duplicate migration version %d", m.version)
			}
			var applied bool
			if err := tx.QueryRow(ctx, "SELECT EXISTS (SELECT FROM schema_migrations WHERE version=$1)", m.version).Scan(&applied); err != nil {
				return err
			}
			if applied {
				continue
			}
			sql, err := fs.ReadFile(dir, m.path)
			if err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, string(sql)); err != nil {
				return fmt.Errorf("%s: %w", m.path, err)
			}
			if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations(version) VALUES ($1)", m.version); err != nil {
				return err
			}
		}
		return nil
	})
}

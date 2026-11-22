package testutil

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/Ender-swe/cueia.io/apps/server/internal/database"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider/providertest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Workspace provisions an isolated database schema and a temporary volume provider.
func Workspace(t *testing.T, mode string) *workspace.Service {
	t.Helper()
	dsn := os.Getenv("CUEIA_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires CUEIA_TEST_DATABASE_URL")
	}
	db, err := pgxpool.New(t.Context(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	schema := "api_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = db.Exec(t.Context(), "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	admin := db
	t.Cleanup(func() {
		if _, err := admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Error(err)
		}
	})
	config := db.Config()
	config.ConnConfig.RuntimeParams["search_path"] = schema
	db, err = pgxpool.NewWithConfig(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	conn, err := db.Acquire(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	err = database.Migrate(t.Context(), conn.Conn(), os.DirFS("../../migrations"))
	conn.Release()
	if err != nil {
		t.Fatal(err)
	}
	return &workspace.Service{DB: db, Provider: &providertest.Fixture{Directory: t.TempDir()}, Deployment: uuid.NewString(), ProviderName: "internal", Region: "local", Mode: mode}
}

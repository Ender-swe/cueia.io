package claim

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestClaimAtVersion(t *testing.T) {
	url := os.Getenv("CUEIA_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("requires migrated CUEIA_TEST_DATABASE_URL")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	if _, err = conn.Exec(ctx, "BEGIN"); err != nil {
		t.Fatal(err)
	}
	defer conn.Exec(ctx, "ROLLBACK")
	var id string
	err = conn.QueryRow(ctx, `WITH w AS (INSERT INTO workspaces DEFAULT VALUES RETURNING id)
		INSERT INTO claim_records(workspace_id,status,version)
		SELECT id,'eligible',4 FROM w RETURNING id`).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range []struct {
		version int64
		success bool
	}{{3, false}, {4, true}, {4, false}, {5, false}} {
		claimed, err := AtVersion(ctx, conn, id, step.version)
		if err != nil || claimed != step.success {
			t.Fatalf("version %d: claimed=%t err=%v", step.version, claimed, err)
		}
	}
	var version int64
	if err = conn.QueryRow(ctx, "SELECT version FROM claim_records WHERE id=$1", id).Scan(&version); err != nil || version != 5 {
		t.Fatalf("persisted version=%d err=%v", version, err)
	}
}

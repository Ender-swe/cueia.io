package database

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Point CUEIA_TEST_DATABASE_URL at an empty, disposable PostgreSQL database.
func TestBootstrapMigration(t *testing.T) {
	url := os.Getenv("CUEIA_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("requires CUEIA_TEST_DATABASE_URL")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		dir = "../../migrations"
	}
	for range 2 {
		if err := Migrate(ctx, conn, os.DirFS(dir)); err != nil {
			t.Fatal(err)
		}
	}
	var versions, tables int
	if err := conn.QueryRow(ctx, "SELECT (SELECT count(*) FROM schema_migrations), count(*) FROM information_schema.tables WHERE table_schema=current_schema()").Scan(&versions, &tables); err != nil {
		t.Fatal(err)
	}
	if versions != 1 || tables != 13 {
		t.Fatalf("got %d versions and %d tables", versions, tables)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	insertID := func(query string, args ...any) string {
		t.Helper()
		var id string
		if err := tx.QueryRow(ctx, query, args...).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	check := func(expected, query string, args ...any) {
		t.Helper()
		sub, err := tx.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		_, err = sub.Exec(ctx, query, args...)
		if expected == "" {
			if err != nil {
				t.Fatal(err)
			}
			if err := sub.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			return
		}
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != expected {
			t.Fatalf("expected SQLSTATE %s, got %v", expected, err)
		}
		if err := sub.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
	}
	a := insertID("INSERT INTO workspaces DEFAULT VALUES RETURNING id")
	b := insertID("INSERT INTO workspaces DEFAULT VALUES RETURNING id")
	const user = "INSERT INTO users(clerk_user_id,workspace_id) VALUES($1,$2)"
	check("", user, "alice", a)
	check("23505", user, "alice", b)
	check("23505", user, "bob", a)
	check("23502", user, nil, b)
	check("23502", user, "bob", nil)
	check("23503", user, "bob", "00000000-0000-0000-0000-000000000000")
	const machine = "INSERT INTO workspace_machines(workspace_id,provider,region,provisioning_state) VALUES($1,'fixture','test','pending')"
	check("", machine, a)
	check("23505", machine, a)
	thread := insertID("INSERT INTO mail_threads(workspace_id,subject) VALUES ($1,'Workshop') RETURNING id", a)
	const message = "INSERT INTO mail_messages(workspace_id,thread_id,sender_address,body_text,sent_at) VALUES($1,$2,'a@example.test','Invitation',now())"
	check("23503", message, b, thread)
	msg := insertID(message+" RETURNING id", a, thread)
	const file = "INSERT INTO files(workspace_id,filename,media_type,byte_size,sha256_digest,storage_key) VALUES ($1,'form.pdf','application/pdf',0,sha256(''::bytea),'form.pdf') RETURNING id"
	local := insertID(file, a)
	foreign := insertID(file, b)
	const attach = "INSERT INTO mail_attachments VALUES ($1,$2,$3)"
	check("", attach, a, msg, local)
	check("23503", attach, a, msg, foreign)
}

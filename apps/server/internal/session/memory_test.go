package session

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Ender-swe/cueia.io/apps/server/internal/database"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session/sessiontest"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider/providertest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestFirstExchangesRecover(t *testing.T) {
	s := testWorkspace(t, "sample")
	f := s.Provider.(*providertest.Fixture)
	ctx := t.Context()
	sub := "first-" + uuid.NewString()
	f.AfterAllocate = func(provider.Machine) error { return errors.New("lost response") }
	issuer := sessiontest.NewIssuer(t)
	t.Setenv("CLERK_ISSUER", issuer.URL)
	t.Setenv("CLERK_AUTHORIZED_PARTY", sessiontest.Origin)
	sessions, err := NewService(t.Context(), s.DB, s.Provision)
	if err != nil {
		t.Fatal(err)
	}
	token := issuer.Token(t, sub)
	if _, err := sessions.Exchange(ctx, token); err == nil {
		t.Fatal("expected interruption")
	}
	f.AfterAllocate = nil
	var pending string
	if err := s.DB.QueryRow(ctx, "SELECT workspace_id FROM users WHERE clerk_user_id=$1", sub).Scan(&pending); err != nil {
		t.Fatal(err)
	}
	// Fail the final seed insert once, after disk publication and before commit.
	schema := "fault_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = s.DB.Exec(ctx, fmt.Sprintf(`CREATE FUNCTION %s() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.workspace_id='%s'::uuid THEN RAISE EXCEPTION 'fixture interruption'; END IF; RETURN NEW; END $$;
 CREATE TRIGGER %s BEFORE INSERT ON workspace_preferences FOR EACH ROW EXECUTE FUNCTION %s()`, schema, pending, schema, schema))
	if err != nil {
		t.Fatal(err)
	}
	undo := fmt.Sprintf("DROP FUNCTION IF EXISTS %s() CASCADE", schema)
	defer s.DB.Exec(context.Background(), undo)
	if _, err = sessions.Exchange(ctx, token); err == nil {
		t.Fatal("expected interrupted transaction")
	}
	if _, err = s.Snapshot(ctx, pending); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatal("partial snapshot", err)
	}
	if _, err = s.DB.Exec(ctx, undo); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	var creds [2]Credential
	var errs [2]error
	var wg sync.WaitGroup
	for i := range creds {
		wg.Go(func() { <-start; creds[i], errs[i] = sessions.Exchange(ctx, token) })
	}
	close(start)
	wg.Wait()
	if errs[0] != nil || errs[1] != nil || creds[0].Token == creds[1].Token {
		t.Fatal("exchanges", errs)
	}
	for _, c := range creds {
		w, err := sessions.Authenticate(ctx, c.Token)
		if err != nil || w != pending {
			t.Fatal(w, err)
		}
	}
	for _, name := range []string{"workspace_machines", "mail_messages", "mail_threads", "calendar_events", "files", "mail_attachments"} {
		var n int
		if err = s.DB.QueryRow(ctx, "SELECT count(*) FROM "+name+" WHERE workspace_id=$1", pending).Scan(&n); err != nil || n != 1 {
			t.Fatal(name, n, err)
		}
	}
	dirs, err := os.ReadDir(f.Directory)
	if err != nil || len(dirs) != 1 {
		t.Fatal("disk allocations", err)
	}
}

func TestStoredSessions(t *testing.T) {
	for _, mode := range []string{"sample", "live"} {
		t.Run(mode, func(t *testing.T) {
			store := testWorkspace(t, mode)
			issuer := sessiontest.NewIssuer(t)
			t.Setenv("CLERK_ISSUER", issuer.URL)
			t.Setenv("CLERK_AUTHORIZED_PARTY", sessiontest.Origin)
			s, err := NewService(t.Context(), store.DB, store.Provision)
			if err != nil {
				t.Fatal(err)
			}
			ctx := t.Context()
			login := func(sub string) Credential {
				t.Helper()
				c, err := s.Exchange(ctx, issuer.Token(t, sub))
				if err != nil {
					t.Fatal(err)
				}
				return c
			}
			sub := uuid.NewString()
			a, b, other := login(sub), login(sub), login(uuid.NewString())
			if a.Token == b.Token {
				t.Fatal("repeated session credential")
			}
			owner := func(c Credential) string {
				t.Helper()
				w, err := s.Authenticate(ctx, c.Token)
				if err != nil {
					t.Fatal(err)
				}
				return w
			}
			if owner(a) != owner(b) || owner(a) == owner(other) {
				t.Fatal("incorrect workspace ownership")
			}
			for _, c := range []Credential{a, b, other} {
				digest := sha256.Sum256([]byte(c.Token))
				var expires time.Time
				if err := store.DB.QueryRow(ctx, "SELECT expires_at FROM sessions WHERE token_digest=$1", digest[:]).Scan(&expires); err != nil || !expires.Equal(c.ExpiresAt) {
					t.Fatal("persisted credential mismatch", err)
				}
			}
			if err := s.Revoke(ctx, a.Token); err != nil {
				t.Fatal(err)
			}
			if _, err := s.Authenticate(ctx, a.Token); !errors.Is(err, pgx.ErrNoRows) {
				t.Fatal("revoked session", err)
			}
			owner(b)
			digest := sha256.Sum256([]byte(b.Token))
			if _, err := store.DB.Exec(ctx, "UPDATE sessions SET expires_at=now()-interval '1 second' WHERE token_digest=$1", digest[:]); err != nil {
				t.Fatal(err)
			}
			if _, err := s.Authenticate(ctx, b.Token); !errors.Is(err, pgx.ErrNoRows) {
				t.Fatal("expired session", err)
			}
			owner(other)
		})
	}
}

func testWorkspace(t *testing.T, mode string) *workspace.Service {
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

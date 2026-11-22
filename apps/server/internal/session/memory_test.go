package session_test

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

	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
	"github.com/Ender-swe/cueia.io/apps/server/internal/testutil"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider/providertest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestFirstExchangesRecover(t *testing.T) {
	s := testutil.Workspace(t, "sample")
	f := s.Provider.(*providertest.Fixture)
	ctx := t.Context()
	sub := "first-" + uuid.NewString()
	f.AfterAllocate = func(provider.Machine) error { return errors.New("lost response") }
	verify := func(_ context.Context, token string) (string, error) { return token, nil }
	sessions := session.NewService(s.DB, verify, s.Provision)
	if _, err := sessions.Exchange(ctx, sub); err == nil {
		t.Fatal("expected interruption")
	}
	f.AfterAllocate = nil
	var pending string
	if err := s.DB.QueryRow(ctx, "SELECT workspace_id FROM users WHERE clerk_user_id=$1", sub).Scan(&pending); err != nil {
		t.Fatal(err)
	}
	// Fail the final seed insert once, after disk publication and before commit.
	schema := "fault_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err := s.DB.Exec(ctx, fmt.Sprintf(`CREATE FUNCTION %s() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.workspace_id='%s'::uuid THEN RAISE EXCEPTION 'fixture interruption'; END IF; RETURN NEW; END $$;
 CREATE TRIGGER %s BEFORE INSERT ON workspace_preferences FOR EACH ROW EXECUTE FUNCTION %s()`, schema, pending, schema, schema))
	if err != nil {
		t.Fatal(err)
	}
	undo := fmt.Sprintf("DROP FUNCTION IF EXISTS %s() CASCADE", schema)
	defer s.DB.Exec(context.Background(), undo)
	if _, err = sessions.Exchange(ctx, sub); err == nil {
		t.Fatal("expected interrupted transaction")
	}
	if _, err = s.Snapshot(ctx, pending); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatal("partial snapshot", err)
	}
	if _, err = s.DB.Exec(ctx, undo); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	var creds [2]session.Credential
	var errs [2]error
	var wg sync.WaitGroup
	for i := range creds {
		wg.Go(func() { <-start; creds[i], errs[i] = sessions.Exchange(ctx, sub) })
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
			store := testutil.Workspace(t, mode)
			s := session.NewService(store.DB, func(_ context.Context, sub string) (string, error) { return sub, nil }, store.Provision)
			ctx := t.Context()
			login := func(sub string) session.Credential {
				t.Helper()
				c, err := s.Exchange(ctx, sub)
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
			owner := func(c session.Credential) string {
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
			for _, c := range []session.Credential{a, b, other} {
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

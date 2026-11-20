package workspace

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Ender-swe/cueia.io/backend/internal/document"
	"github.com/Ender-swe/cueia.io/backend/internal/session"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/Ender-swe/cueia.io/backend/internal/database"
	"github.com/Ender-swe/cueia.io/backend/internal/workspace/provider"
	"github.com/Ender-swe/cueia.io/backend/internal/workspace/provider/providertest"
	"github.com/Ender-swe/cueia.io/backend/internal/workspace/sample"
	"github.com/Ender-swe/cueia.io/backend/internal/workspace/volume"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testService(t *testing.T) *Service {
	t.Helper()
	url := os.Getenv("CUEIA_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("requires CUEIA_TEST_DATABASE_URL")
	}
	ctx := t.Context()
	db, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	c, err := db.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = database.Migrate(ctx, c.Conn(), os.DirFS("../../migrations"))
	c.Release()
	if err != nil {
		t.Fatal(err)
	}
	return &Service{db, &providertest.Fixture{Directory: t.TempDir()}, "test-" + uuid.NewString(), "internal", "local", "sample"}
}
func TestPublicationRecovery(t *testing.T) {
	s := testService(t)
	ctx := t.Context()
	w, err := s.Provision(ctx, uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	disk, err := s.open(ctx, w)
	if err != nil {
		t.Fatal(err)
	}
	defer disk.Close()
	d := disk.(*volume.Store)
	if err = d.WriteFile("keep.txt", []byte("ordinary"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, point := range []string{"staging", "published", "metadata"} {
		t.Run(point, func(t *testing.T) {
			name := point + ".pdf"
			d.Interrupt = func(boundary string) error {
				if point == boundary {
					return errors.New("interrupted")
				}
				return nil
			}
			err := pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
				_, err := publish(ctx, tx, w, d, name, sample.PDF)
				if err == nil {
					return errors.New("simulate rollback before commit")
				}
				return err
			})
			if err == nil {
				t.Fatal("interruption returned success")
			}
			var n int
			if err = s.DB.QueryRow(ctx, "SELECT count(*) FROM files WHERE workspace_id=$1 AND filename=$2", w, name).Scan(&n); err != nil || n != 0 {
				t.Fatal(n, err)
			}
			entries, err := s.Files(ctx, w, ".", name)
			if err != nil || len(entries) != 0 {
				t.Fatal(entries, err)
			}
			d.Interrupt = nil
			if err = d.Reconcile(); err != nil {
				t.Fatal(err)
			}
			var first string
			for range 2 {
				err = pgx.BeginFunc(ctx, s.DB, func(tx pgx.Tx) error {
					id, err := publish(ctx, tx, w, d, name, sample.PDF)
					if first != "" && id != first {
						return fmt.Errorf("file ID changed")
					}
					first = id
					return err
				})
				if err != nil {
					t.Fatal(err)
				}
			}
			got, err := s.Document(ctx, w, first)
			if err != nil || !bytes.Equal(got, sample.PDF) {
				t.Fatal("recovered bytes", err)
			}
		})
	}
	got, err := fs.ReadFile(d, "keep.txt")
	if err != nil || string(got) != "ordinary" {
		t.Fatal("ordinary file lost", err)
	}
}
func TestStorageIsolation(t *testing.T) {
	s := testService(t)
	ctx := t.Context()
	alice, err := s.Provision(ctx, uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	bob, err := s.Provision(ctx, uuid.NewString())
	if err != nil || alice == bob {
		t.Fatal(bob, err)
	}
	var fileID string
	for _, w := range []string{alice, bob} {
		d, err := s.open(ctx, w)
		if err != nil {
			t.Fatal(err)
		}
		local := d.(*volume.Store)
		if err = local.Mkdir("reports", 0700); err != nil {
			t.Fatal(err)
		}
		if err = d.WriteFile("reports/activity.csv", []byte(w), 0600); err != nil {
			t.Fatal(err)
		}
		if err = d.WriteFile("reports/form.pdf", sample.PDF, 0600); err != nil {
			t.Fatal(err)
		}
		outside := filepath.Join(t.TempDir(), "foreign")
		if err = os.WriteFile(outside, []byte("foreign"), 0600); err != nil {
			t.Fatal(err)
		}
		if err = os.Symlink(outside, filepath.Join(local.Name(), "escape")); err != nil {
			t.Fatal(err)
		}
		for _, p := range []string{"escape", "../foreign"} {
			if _, err := s.Read(ctx, w, p); err == nil {
				t.Fatal("escaped", p)
			}
		}
		entries, err := s.Files(ctx, w, ".", "activity")
		if err != nil || len(entries) != 1 || entries[0].Path != "reports/activity.csv" {
			t.Fatal(entries, err)
		}
		got, err := s.Read(ctx, w, "reports/activity.csv")
		if err != nil || string(got) != w {
			t.Fatal(err)
		}
		var results [2]string
		var errs [2]error
		var wg sync.WaitGroup
		for i := range results {
			wg.Go(func() { results[i], errs[i] = s.Register(ctx, w, "reports/form.pdf") })
		}
		wg.Wait()
		if errs[0] != nil || errs[1] != nil || results[0] != results[1] {
			t.Fatal(results, errs)
		}
		if w == alice {
			fileID = results[0]
		} else if results[0] == fileID {
			t.Fatal("shared file ID")
		}
		if err = d.WriteFile("reports/form.pdf", append(bytes.Clone(sample.PDF), '\n'), 0600); err != nil {
			t.Fatal(err)
		}
		newer, err := s.Register(ctx, w, "reports/form.pdf")
		if err != nil || newer == results[0] {
			t.Fatal("unchanged version", err)
		}
		got, err = s.Document(ctx, w, results[0])
		if err != nil || !bytes.Equal(got, sample.PDF) {
			t.Fatal("old bytes changed", err)
		}
		d.Close()
	}
	if _, err = s.Document(ctx, bob, fileID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatal("foreign read", err)
	}
	if _, err = s.Preferences(ctx, alice, map[string]string{"timeFormat": "24h"}); err != nil {
		t.Fatal(err)
	}
	for _, w := range []string{alice, bob} {
		snap, err := s.Snapshot(ctx, w)
		if err != nil {
			t.Fatal(err)
		}
		if (snap.Preferences["timeFormat"] == "24h") != (w == alice) {
			t.Fatal("shared preference")
		}
	}
	var m provider.Machine
	if err = s.DB.QueryRow(ctx, "SELECT vm_resource_id,workspace_disk_resource_id FROM workspace_machines WHERE workspace_id=$1", alice).Scan(&m.ID, &m.DiskID); err != nil {
		t.Fatal(err)
	}
	_, err = s.Provider.Replace(ctx, provider.Owner{Deployment: s.Deployment, Workspace: alice}, m)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Files(ctx, alice, ".", ""); err == nil {
		t.Fatal("unavailable binding reported as available")
	}
}

func TestFirstExchangesRecover(t *testing.T) {
	s := testService(t)
	f := s.Provider.(*providertest.Fixture)
	ctx := t.Context()
	sub := "first-" + uuid.NewString()
	f.AfterAllocate = func(provider.Machine) error { return errors.New("lost response") }
	verify := func(_ context.Context, token string) (string, error) { return token, nil }
	sessions := session.Service{DB: s.DB, Provision: s.Provision, Verify: verify}
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
	live := testService(t)
	isolated := "live_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = live.DB.Exec(ctx, "CREATE SCHEMA "+isolated)
	if err != nil {
		t.Fatal(err)
	}
	defer live.DB.Exec(context.Background(), "DROP SCHEMA "+isolated+" CASCADE")
	cfg := live.DB.Config()
	cfg.ConnConfig.RuntimeParams["search_path"] = isolated
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	c, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = database.Migrate(ctx, c.Conn(), os.DirFS("../../migrations"))
	c.Release()
	if err != nil {
		t.Fatal(err)
	}
	live.DB = pool
	live.Mode = "live"
	live.Provider = f
	w, err := live.Provision(ctx, sub)
	if err != nil || w == pending {
		t.Fatal("live provisioning", err)
	}
	v, err := live.Snapshot(ctx, w)
	if err != nil || string(v.Threads) != "[]" || len(v.Preferences) != 0 {
		t.Fatal("live seed", err)
	}
	dirs, err = os.ReadDir(f.Directory)
	if err != nil || len(dirs) != 2 {
		t.Fatal("live disk", err)
	}
}

func TestOwnerCollectionsAndSessions(t *testing.T) {
	s := testService(t)
	ctx := t.Context()
	sessions := session.Service{DB: s.DB, Verify: func(_ context.Context, sub string) (string, error) { return sub, nil }, Provision: s.Provision}
	sub := uuid.NewString()
	login := func(sub string) session.Credential {
		t.Helper()
		c, err := sessions.Exchange(ctx, sub)
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	a, b, other := login(sub), login(sub), login(uuid.NewString())
	owner := func(c session.Credential) string {
		t.Helper()
		w, err := sessions.Authenticate(ctx, c.Token)
		if err != nil {
			t.Fatal(err)
		}
		return w
	}
	alice, bob := owner(a), owner(other)
	if _, err := s.Preferences(ctx, alice, map[string]string{"timeFormat": "24h"}); err != nil {
		t.Fatal(err)
	}
	snap, err := s.Snapshot(ctx, owner(b))
	if err != nil || snap.Preferences["timeFormat"] != "24h" {
		t.Fatal("other session did not see preference", err)
	}
	var threads, events, fileIDs []string
	for _, w := range []string{alice, bob} {
		tt, err := s.Threads(ctx, w)
		if err != nil {
			t.Fatal(err)
		}
		ee, err := s.Events(ctx, w)
		if err != nil {
			t.Fatal(err)
		}
		var th []struct {
			ID       string
			Messages []struct{ Attachments []struct{ FileID string } }
		}
		var ev []struct{ ID string }
		if err = json.Unmarshal(tt, &th); err != nil || len(th) != 1 || len(th[0].Messages) != 1 || len(th[0].Messages[0].Attachments) != 1 {
			t.Fatal("threads", err)
		}
		if err = json.Unmarshal(ee, &ev); err != nil || len(ev) != 1 {
			t.Fatal("events", err)
		}
		threads = append(threads, th[0].ID)
		events = append(events, ev[0].ID)
		fileIDs = append(fileIDs, th[0].Messages[0].Attachments[0].FileID)
	}
	if threads[0] == threads[1] || events[0] == events[1] || fileIDs[0] == fileIDs[1] {
		t.Fatal("owner lists share a record")
	}
	if _, err = s.Document(ctx, bob, fileIDs[0]); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatal("foreign file accessible", err)
	}
	got, err := s.Document(ctx, alice, fileIDs[0])
	if err != nil || !bytes.Equal(got, sample.PDF) {
		t.Fatal("Alice's original bytes", err)
	}
	snap, err = s.Snapshot(ctx, bob)
	if err != nil || snap.Preferences["timeFormat"] != "12h" {
		t.Fatal("Bob's default preference", err)
	}
}

func TestDirectWritesAndPreview(t *testing.T) {
	s := testService(t)
	ctx := t.Context()
	w, err := s.Provision(ctx, uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	disk, err := s.open(ctx, w)
	if err != nil {
		t.Fatal(err)
	}
	root := disk.(*volume.Store).Name()
	disk.Close()
	countFiles := func() int {
		t.Helper()
		var n int
		if err := s.DB.QueryRow(ctx, "SELECT count(*) FROM files WHERE workspace_id=$1", w).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	before := countFiles()
	cmd := exec.CommandContext(ctx, "sh", "-c", "mkdir -p reports/sub && echo workshop > reports/owner.txt")
	cmd.Dir = root
	if err = cmd.Run(); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, "reports/sub/agent.csv"), []byte("entry,count\nforms,2\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"reports/owner.txt", "reports/sub/agent.csv"} {
		entries, err := s.Files(ctx, w, path.Dir(name), "")
		if err != nil {
			t.Fatal(err)
		}
		if !slices.ContainsFunc(entries, func(e Entry) bool { return e.Path == name }) {
			t.Fatal("refresh missed", name)
		}
		entries, err = s.Files(ctx, w, ".", path.Base(name))
		if err != nil || len(entries) != 1 || entries[0].Path != name {
			t.Fatal("search", entries, err)
		}
	}
	if countFiles() != before {
		t.Fatal("direct writes registered as documents")
	}
	if err = os.WriteFile(filepath.Join(root, "Attachments"), []byte("ordinary collision"), 0600); err != nil {
		t.Fatal(err)
	}
	ordinary, err := s.Read(ctx, w, "Attachments")
	if err != nil || string(ordinary) != "ordinary collision" {
		t.Fatal("ordinary entry hidden", err)
	}
	entries, err := s.Files(ctx, w, ".", "north-pier-registration.pdf")
	if err != nil || len(entries) != 1 || path.Dir(entries[0].Path) != "Attachments (2)" {
		t.Fatal("seed entry hidden", entries, err)
	}
	filename := "reports/form.pdf"
	if err = os.WriteFile(filepath.Join(root, filename), sample.PDF, 0600); err != nil {
		t.Fatal(err)
	}
	id, err := s.Register(ctx, w, filename)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := document.Load(ctx, s.DB)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Owns(ctx, w, id); err != nil {
		t.Fatal(err)
	}
	link := signer.Issue(w, id)
	changed := append(bytes.Clone(sample.PDF), '\n')
	if err = os.WriteFile(filepath.Join(root, filename), changed, 0600); err != nil {
		t.Fatal(err)
	}
	next, err := s.Register(ctx, w, filename)
	if err != nil || next == id {
		t.Fatal("expected new version", err)
	}
	u, err := url.Parse(link.URL)
	if err != nil {
		t.Fatal(err)
	}
	authorizedOwner, err := signer.Redeem(path.Base(u.Path), "GET", u.Query().Get("token"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Document(ctx, authorizedOwner, id)
	if err != nil || !bytes.Equal(got, sample.PDF) {
		t.Fatal("earlier preview changed", err)
	}
	got, err = s.Document(ctx, w, next)
	if err != nil || !bytes.Equal(got, changed) {
		t.Fatal("new version", err)
	}
}

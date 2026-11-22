package workspace

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	"github.com/Ender-swe/cueia.io/apps/server/internal/database"
	"github.com/Ender-swe/cueia.io/apps/server/internal/document"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider/providertest"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/sample"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/volume"
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
				_, err := publish(ctx, tx, w, d, name, "application/pdf", sample.PDF)
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
					id, err := publish(ctx, tx, w, d, name, "application/pdf", sample.PDF)
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
			got, err := s.ReadFileVersion(ctx, w, first)
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
		if err = d.WriteFile("reports/invalid.pdf", []byte("plain text"), 0600); err != nil {
			t.Fatal(err)
		}
		documents := document.NewService(s, document.Signer{})
		if _, err := documents.Register(ctx, w, "reports/invalid.pdf"); !errors.Is(err, fs.ErrInvalid) {
			t.Fatal("expected invalid PDF error", err)
		}
		csvID, err := s.RegisterFileVersion(ctx, w, "reports/activity.csv", "text/csv", nil)
		if err != nil {
			t.Fatal(err)
		}
		var mediaType string
		if err := s.DB.QueryRow(ctx, "SELECT media_type FROM files WHERE workspace_id=$1 AND id=$2", w, csvID).Scan(&mediaType); err != nil || mediaType != "text/csv" {
			t.Fatal("registered CSV media type", mediaType, err)
		}
		var results [2]string
		var errs [2]error
		var wg sync.WaitGroup
		for i := range results {
			wg.Go(func() {
				results[i], errs[i] = s.RegisterFileVersion(ctx, w, "reports/form.pdf", "application/pdf", nil)
			})
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
		newer, err := s.RegisterFileVersion(ctx, w, "reports/form.pdf", "application/pdf", nil)
		if err != nil || newer == results[0] {
			t.Fatal("unchanged version", err)
		}
		got, err = s.ReadFileVersion(ctx, w, results[0])
		if err != nil || !bytes.Equal(got, sample.PDF) {
			t.Fatal("old bytes changed", err)
		}
		d.Close()
	}
	if _, err = s.ReadFileVersion(ctx, bob, fileID); !errors.Is(err, pgx.ErrNoRows) {
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
	if got, err := s.ReadFileVersion(ctx, alice, fileID); err != nil || !bytes.Equal(got, sample.PDF) {
		t.Fatal("replacement did not reopen captured PDF", err)
	}
}

func TestLiveProvisioningUsesSeparateDisk(t *testing.T) {
	ctx := t.Context()
	sampleStore := testService(t)
	sub := uuid.NewString()
	pending, err := sampleStore.Provision(ctx, sub)
	if err != nil {
		t.Fatal(err)
	}
	f := sampleStore.Provider.(*providertest.Fixture)
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
	dirs, err := os.ReadDir(f.Directory)
	if err != nil || len(dirs) != 2 {
		t.Fatal("live disk", err)
	}
}

func TestOwnerCollections(t *testing.T) {
	s := testService(t)
	ctx := t.Context()
	provision := func() string {
		t.Helper()
		owner, err := s.Provision(ctx, uuid.NewString())
		if err != nil {
			t.Fatal(err)
		}
		return owner
	}
	alice, bob := provision(), provision()
	if _, err := s.Preferences(ctx, alice, map[string]string{"timeFormat": "24h"}); err != nil {
		t.Fatal(err)
	}
	snap, err := s.Snapshot(ctx, alice)
	if err != nil || snap.Preferences["timeFormat"] != "24h" {
		t.Fatal("owner preference missing", err)
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
		attachmentID := th[0].Messages[0].Attachments[0].FileID
		registeredID, err := s.RegisterFileVersion(ctx, w, "Attachments/north-pier-registration.pdf", "application/pdf", nil)
		if err != nil || registeredID != attachmentID {
			t.Fatal("registration did not reuse the attachment ID", registeredID, attachmentID, err)
		}
		fileIDs = append(fileIDs, attachmentID)
	}
	if threads[0] == threads[1] || events[0] == events[1] || fileIDs[0] == fileIDs[1] {
		t.Fatal("owner lists share a record")
	}
	if _, err = s.ReadFileVersion(ctx, bob, fileIDs[0]); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatal("foreign file accessible", err)
	}
	got, err := s.ReadFileVersion(ctx, alice, fileIDs[0])
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
	id, err := s.RegisterFileVersion(ctx, w, filename, "application/pdf", nil)
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
	next, err := s.RegisterFileVersion(ctx, w, filename, "application/pdf", nil)
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
	got, err := s.ReadFileVersion(ctx, authorizedOwner, id)
	if err != nil || !bytes.Equal(got, sample.PDF) {
		t.Fatal("earlier preview changed", err)
	}
	got, err = s.ReadFileVersion(ctx, w, next)
	if err != nil || !bytes.Equal(got, changed) {
		t.Fatal("new version", err)
	}
}

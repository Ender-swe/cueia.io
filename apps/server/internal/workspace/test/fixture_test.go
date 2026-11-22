package workspace_test

import (
	"errors"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider/providertest"
	"io/fs"
	"testing"

	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider"
)

func TestAllocationAndReplacement(t *testing.T) {
	ctx := t.Context()
	f := &providertest.Fixture{Directory: t.TempDir()}
	a := provider.Owner{Deployment: "sample", Workspace: "alice"}
	b := provider.Owner{Deployment: "sample", Workspace: "bob"}
	var created provider.Machine
	f.AfterAllocate = func(m provider.Machine) error { created = m; return errors.New("lost response") }
	if _, err := f.FindOrAllocate(ctx, a); err == nil {
		t.Fatal("expected lost response")
	}
	f = &providertest.Fixture{Directory: f.Directory}
	ma, err := f.FindOrAllocate(ctx, a)
	if err != nil || ma != created {
		t.Fatalf("retry: %v %v", ma, err)
	}
	mb, err := f.FindOrAllocate(ctx, b)
	if err != nil || ma.ID == mb.ID || ma.DiskID == mb.DiskID {
		t.Fatal("invalid allocation", err)
	}
	for _, p := range []struct {
		owner   provider.Owner
		machine provider.Machine
		other   provider.Machine
	}{{a, ma, mb}, {b, mb, ma}} {
		if _, err := f.Open(ctx, p.owner, p.other); !errors.Is(err, fs.ErrPermission) {
			t.Fatal("foreign mount", err)
		}
		if _, err := f.Replace(ctx, p.owner, p.other); !errors.Is(err, fs.ErrPermission) {
			t.Fatal("foreign replacement", err)
		}
		d, err := f.Open(ctx, p.owner, p.machine)
		if err != nil {
			t.Fatal(err)
		}
		if err = d.WriteFile("notes", []byte(p.owner.Workspace), 0600); err != nil {
			t.Fatal(err)
		}
		d.Close()
	}
	replacement, err := f.Replace(ctx, a, ma)
	if err != nil || replacement.ID == ma.ID || replacement.DiskID != ma.DiskID {
		t.Fatal("replacement", err)
	}
	for _, p := range []struct {
		owner   provider.Owner
		machine provider.Machine
	}{{a, replacement}, {b, mb}} {
		d, err := f.Open(ctx, p.owner, p.machine)
		if err != nil {
			t.Fatal(err)
		}
		bytes, err := fs.ReadFile(d, "notes")
		d.Close()
		if err != nil || string(bytes) != p.owner.Workspace {
			t.Fatal("incorrect disk contents", err)
		}
	}
}

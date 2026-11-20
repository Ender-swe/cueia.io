package providertest

import (
	"os"
	"testing"

	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider"
)

// TestReplaceAcceptanceMachine operates on the internal provider's disposable
// disk directory while the sample API is stopped during the recovery check.
func TestReplaceAcceptanceMachine(t *testing.T) {
	dir := os.Getenv("CUEIA_TEST_DISK_DIR")
	workspace := os.Getenv("CUEIA_TEST_REPLACE_WORKSPACE")
	if dir == "" || workspace == "" {
		t.Skip("requires a disposable workspace to replace")
	}
	f := Fixture{Directory: dir}
	owner := provider.Owner{Deployment: os.Getenv("COMPOSE_PROJECT_NAME"), Workspace: workspace}
	allocation, err := f.ownerDir(owner)
	if err != nil {
		t.Fatal(err)
	}
	old, err := readBinding(allocation)
	if err != nil {
		t.Fatal(err)
	}
	next, err := f.Replace(t.Context(), owner, old)
	if err != nil || next.ID == old.ID || next.DiskID != old.DiskID {
		t.Fatal("workspace disk did not survive replacement", err)
	}
}

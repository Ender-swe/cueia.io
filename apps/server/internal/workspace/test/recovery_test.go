package workspace_test

import (
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider/providertest"
	"testing"

	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider"
)

// TestReplaceAcceptanceMachine operates on the internal provider's disposable
// disk directory while the sample API is stopped during the recovery check.
func TestReplaceAcceptanceMachine(t *testing.T) {
	settings := loadRecoveryConfig()
	if settings.diskDir == "" || settings.workspace == "" {
		t.Skip("requires a disposable workspace to replace")
	}
	f := providertest.Fixture{Directory: settings.diskDir}
	owner := provider.Owner{Deployment: settings.deployment, Workspace: settings.workspace}
	old, err := f.Lookup(owner)
	if err != nil {
		t.Fatal(err)
	}
	next, err := f.Replace(t.Context(), owner, old)
	if err != nil || next.ID == old.ID || next.DiskID != old.DiskID {
		t.Fatal("workspace disk did not survive replacement", err)
	}
}

package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Ender-swe/cueia.io/apps/server/internal/document"
	"github.com/Ender-swe/cueia.io/apps/server/internal/middleware"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider/providertest"
)

// TestBrowserJourney serves the production dashboard and API on one loopback
// origin. The developer signs into the existing Clerk test instance in Chromium.
func TestBrowserJourney(t *testing.T) {
	issuer := os.Getenv("CUEIA_BROWSER_CLERK_ISSUER")
	if issuer == "" {
		t.Skip("requires an interactive Clerk test login")
	}
	store := testWorkspace(t, "sample")
	dist, err := filepath.Abs("../../../dashboard/dist")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(dist, "index.html")); err != nil {
		t.Fatal("build the dashboard before running the browser journey")
	}
	server := httptest.NewUnstartedServer(nil)
	origin := "http://" + server.Listener.Addr().String()
	s := session.Service{DB: store.DB, Verify: session.Clerk(t.Context(), issuer, origin), Provision: store.Provision}
	mux := routes(s.Authenticate, nil)
	workspaceRoutes(mux, s, store)
	signer, err := document.Load(t.Context(), store.DB)
	if err != nil {
		t.Fatal(err)
	}
	previewRoutes(mux, s, store, signer)
	root := http.NewServeMux()
	root.Handle("/api/", middleware.Protect(mux, origin, middleware.NewLimiter()))
	for _, path := range []string{"/workspace", "/sign-in", "/sign-in/", "/sign-up", "/sign-up/"} {
		root.HandleFunc("GET "+path, func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, filepath.Join(dist, "index.html")) })
	}
	root.Handle("/", http.FileServer(http.Dir(dist)))
	server.Config.Handler = root
	server.Start()
	defer server.Close()
	cmd := exec.CommandContext(t.Context(), "bun", "run", "journey.ts")
	cmd.Dir = "../../../browser-worker"
	cmd.Env = append(os.Environ(), "CUEIA_DASHBOARD_ORIGIN="+origin, "CUEIA_TEST_DEPLOYMENT="+store.Deployment, "CUEIA_TEST_DISKS="+store.Provider.(*providertest.Fixture).Directory)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatal("browser journey failed", err)
	}
}

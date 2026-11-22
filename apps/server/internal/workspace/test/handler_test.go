package workspace_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace"
)

func TestWorkspaceRoutesRequireSession(t *testing.T) {
	mux := http.NewServeMux()
	workspace.NewHandler(nil, func(context.Context, string) (string, error) { return "", session.ErrUnauthorized }).Register(mux)
	for _, route := range []struct{ method, path string }{
		{"GET", "/api/workspace"}, {"GET", "/api/mail/threads"}, {"GET", "/api/calendar/events"},
		{"PATCH", "/api/workspace/preferences"}, {"GET", "/api/files"}, {"GET", "/api/files/search"}, {"GET", "/api/files/content"},
	} {
		t.Run(route.path, func(t *testing.T) {
			for _, token := range []string{"", "invalid"} {
				req := httptest.NewRequest(route.method, route.path, nil)
				if token != "" {
					req.Header.Set("Authorization", "Bearer "+token)
				}
				res := httptest.NewRecorder()
				mux.ServeHTTP(res, req)
				if res.Code != http.StatusUnauthorized || res.Header().Get("Cache-Control") != "no-store" {
					t.Fatalf("token %q: status %d, headers %v", token, res.Code, res.Header())
				}
			}
		})
	}
}

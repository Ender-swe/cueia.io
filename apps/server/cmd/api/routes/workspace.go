package routes

import (
	"context"
	"net/http"

	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace"
)

func registerWorkspace(mux *http.ServeMux, authenticate func(context.Context, string) (string, error), h *workspace.Handler) {
	mux.HandleFunc("GET /api/workspace", session.Require(authenticate, h.Snapshot))
	mux.HandleFunc("GET /api/mail/threads", session.Require(authenticate, h.Threads))
	mux.HandleFunc("GET /api/calendar/events", session.Require(authenticate, h.Events))
	mux.HandleFunc("PATCH /api/workspace/preferences", session.Require(authenticate, h.Preferences))
	mux.HandleFunc("GET /api/files", session.Require(authenticate, h.Files))
	mux.HandleFunc("GET /api/files/search", session.Require(authenticate, h.SearchFiles))
	mux.HandleFunc("GET /api/files/content", session.Require(authenticate, h.ReadFile))
}

package routes

import (
	"context"
	"net/http"

	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace"
)

func RegisterWorkspace(mux *http.ServeMux, authenticate func(context.Context, string) (string, error), h *workspace.Handler) {
	mux.HandleFunc("GET /workspace", session.Require(authenticate, h.Snapshot))
	mux.HandleFunc("GET /mail/threads", session.Require(authenticate, h.Threads))
	mux.HandleFunc("GET /calendar/events", session.Require(authenticate, h.Events))
	mux.HandleFunc("PATCH /workspace/preferences", session.Require(authenticate, h.Preferences))
	mux.HandleFunc("GET /files", session.Require(authenticate, h.Files))
	mux.HandleFunc("GET /files/search", session.Require(authenticate, h.SearchFiles))
	mux.HandleFunc("GET /files/content", session.Require(authenticate, h.ReadFile))
}

package main

import (
	"net/http"

	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace"
)

func workspaceRoutes(mux *http.ServeMux, sessions *session.Service, store *workspace.Service) {
	session.NewHandler(sessions).Register(mux)
	workspace.NewHandler(store, sessions.Authenticate).Register(mux)
}

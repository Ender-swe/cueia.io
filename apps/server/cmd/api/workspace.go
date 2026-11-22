package main

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/Ender-swe/cueia.io/apps/server/internal/middleware"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace"
)

func workspaceRoutes(mux *http.ServeMux, sessions *session.Service, store *workspace.Service) {
	private := func(pattern string, handler session.OwnerHandler) {
		mux.HandleFunc(pattern, session.Require(sessions.Authenticate, handler))
	}
	session.NewHandler(sessions).Register(mux)
	private("GET /api/workspace", func(w http.ResponseWriter, r *http.Request, owner string) {
		v, err := store.Snapshot(r.Context(), owner)
		middleware.Result(w, v, err)
	})
	private("GET /api/mail/threads", func(w http.ResponseWriter, r *http.Request, owner string) {
		v, err := store.Threads(r.Context(), owner)
		middleware.Result(w, v, err)
	})
	private("GET /api/calendar/events", func(w http.ResponseWriter, r *http.Request, owner string) {
		v, err := store.Events(r.Context(), owner)
		middleware.Result(w, v, err)
	})
	private("PATCH /api/workspace/preferences", func(w http.ResponseWriter, r *http.Request, owner string) {
		var input map[string]*string
		if middleware.Decode(w, r, &input) != nil || input == nil {
			middleware.RespondError(w, fs.ErrInvalid)
			return
		}
		changes := make(map[string]string, len(input))
		for k, v := range input {
			if v == nil {
				middleware.RespondError(w, fs.ErrInvalid)
				return
			}
			changes[k] = *v
		}
		v, err := store.Preferences(r.Context(), owner, changes)
		middleware.Result(w, v, err)
	})
	private("GET /api/files", func(w http.ResponseWriter, r *http.Request, owner string) {
		dir := r.URL.Query().Get("directory")
		if dir != "" && !fs.ValidPath(dir) {
			middleware.RespondError(w, fs.ErrInvalid)
			return
		}
		v, err := store.Files(r.Context(), owner, dir, "")
		middleware.Result(w, v, err)
	})
	private("GET /api/files/search", func(w http.ResponseWriter, r *http.Request, owner string) {
		name := r.URL.Query().Get("name")
		if strings.TrimSpace(name) == "" {
			middleware.RespondError(w, fs.ErrInvalid)
			return
		}
		v, err := store.Files(r.Context(), owner, ".", name)
		middleware.Result(w, v, err)
	})
	private("GET /api/files/content", func(w http.ResponseWriter, r *http.Request, owner string) {
		v, err := store.Read(r.Context(), owner, r.URL.Query().Get("path"))
		middleware.Content(w, "application/octet-stream", v, err)
	})
}

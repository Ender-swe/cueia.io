package main

import (
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/Ender-swe/cueia.io/backend/internal/session"
	"github.com/Ender-swe/cueia.io/backend/internal/workspace"
	"github.com/google/uuid"
)

func decode(w http.ResponseWriter, r *http.Request, value any) error {
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	d.DisallowUnknownFields()
	if d.Decode(value) != nil || d.Decode(new(any)) != io.EOF {
		return fs.ErrInvalid
	}
	return nil
}

func result[T any](w http.ResponseWriter, value T, err error) {
	if err != nil {
		respondError(w, err)
		return
	}
	writeJSON(w, 200, value)
}
func content(w http.ResponseWriter, kind string, data []byte, err error) {
	if err != nil {
		respondError(w, err)
		return
	}
	w.Header().Set("Content-Type", kind)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(data)
}

func workspaceRoutes(mux *http.ServeMux, sessions session.Service, store *workspace.Service) {
	private := func(pattern string, handler ownerHandler) {
		mux.HandleFunc(pattern, requireSession(sessions.Authenticate, handler))
	}
	mux.HandleFunc("POST /api/session", sessions.ExchangeHTTP)
	private("DELETE /api/session", func(w http.ResponseWriter, r *http.Request, _ string) {
		if err := sessions.Revoke(r.Context(), session.Bearer(r)); err != nil {
			respondError(w, err)
			return
		}
		w.WriteHeader(204)
	})
	private("GET /api/workspace", func(w http.ResponseWriter, r *http.Request, owner string) {
		v, err := store.Snapshot(r.Context(), owner)
		result(w, v, err)
	})
	private("GET /api/mail/threads", func(w http.ResponseWriter, r *http.Request, owner string) {
		v, err := store.Threads(r.Context(), owner)
		result(w, v, err)
	})
	private("GET /api/calendar/events", func(w http.ResponseWriter, r *http.Request, owner string) {
		v, err := store.Events(r.Context(), owner)
		result(w, v, err)
	})
	private("PATCH /api/workspace/preferences", func(w http.ResponseWriter, r *http.Request, owner string) {
		var input map[string]*string
		if decode(w, r, &input) != nil || input == nil {
			respondError(w, fs.ErrInvalid)
			return
		}
		changes := make(map[string]string, len(input))
		for k, v := range input {
			if v == nil {
				respondError(w, fs.ErrInvalid)
				return
			}
			changes[k] = *v
		}
		v, err := store.Preferences(r.Context(), owner, changes)
		result(w, v, err)
	})
	private("GET /api/files", func(w http.ResponseWriter, r *http.Request, owner string) {
		dir := r.URL.Query().Get("directory")
		if dir != "" && !fs.ValidPath(dir) {
			respondError(w, fs.ErrInvalid)
			return
		}
		v, err := store.Files(r.Context(), owner, dir, "")
		result(w, v, err)
	})
	private("GET /api/files/search", func(w http.ResponseWriter, r *http.Request, owner string) {
		name := r.URL.Query().Get("name")
		if strings.TrimSpace(name) == "" {
			respondError(w, fs.ErrInvalid)
			return
		}
		v, err := store.Files(r.Context(), owner, ".", name)
		result(w, v, err)
	})
	private("GET /api/files/content", func(w http.ResponseWriter, r *http.Request, owner string) {
		v, err := store.Read(r.Context(), owner, r.URL.Query().Get("path"))
		content(w, "application/octet-stream", v, err)
	})
	private("POST /api/documents", func(w http.ResponseWriter, r *http.Request, owner string) {
		var input struct {
			Path string `json:"path"`
		}
		if decode(w, r, &input) != nil {
			respondError(w, fs.ErrInvalid)
			return
		}
		id, err := store.Register(r.Context(), owner, input.Path)
		result(w, map[string]string{"fileId": id}, err)
	})
	private("GET /api/documents/{fileId}/content", func(w http.ResponseWriter, r *http.Request, owner string) {
		id := r.PathValue("fileId")
		if uuid.Validate(id) != nil {
			respondError(w, fs.ErrInvalid)
			return
		}
		v, err := store.Document(r.Context(), owner, id)
		content(w, "application/pdf", v, err)
	})
}

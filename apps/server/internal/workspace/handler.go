package workspace

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/Ender-swe/cueia.io/apps/server/internal/middleware"
)

type Handler struct {
	store *Service
}

func NewHandler(store *Service) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Snapshot(w http.ResponseWriter, r *http.Request, owner string) {
	v, err := h.store.Snapshot(r.Context(), owner)
	middleware.Result(w, v, err)
}

func (h *Handler) Threads(w http.ResponseWriter, r *http.Request, owner string) {
	v, err := h.store.Threads(r.Context(), owner)
	middleware.Result(w, v, err)
}

func (h *Handler) Events(w http.ResponseWriter, r *http.Request, owner string) {
	v, err := h.store.Events(r.Context(), owner)
	middleware.Result(w, v, err)
}

func (h *Handler) Preferences(w http.ResponseWriter, r *http.Request, owner string) {
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
	v, err := h.store.Preferences(r.Context(), owner, changes)
	middleware.Result(w, v, err)
}

func (h *Handler) Files(w http.ResponseWriter, r *http.Request, owner string) {
	dir := r.URL.Query().Get("directory")
	if dir != "" && !fs.ValidPath(dir) {
		middleware.RespondError(w, fs.ErrInvalid)
		return
	}
	v, err := h.store.Files(r.Context(), owner, dir, "")
	middleware.Result(w, v, err)
}

func (h *Handler) SearchFiles(w http.ResponseWriter, r *http.Request, owner string) {
	name := r.URL.Query().Get("name")
	if strings.TrimSpace(name) == "" {
		middleware.RespondError(w, fs.ErrInvalid)
		return
	}
	v, err := h.store.Files(r.Context(), owner, ".", name)
	middleware.Result(w, v, err)
}

func (h *Handler) ReadFile(w http.ResponseWriter, r *http.Request, owner string) {
	v, err := h.store.Read(r.Context(), owner, r.URL.Query().Get("path"))
	middleware.Content(w, "application/octet-stream", v, err)
}

package document

import (
	"context"
	"io/fs"
	"net/http"

	"github.com/Ender-swe/cueia.io/apps/server/internal/middleware"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
	"github.com/google/uuid"
)

type Store interface {
	Register(context.Context, string, string) (string, error)
	Document(context.Context, string, string) ([]byte, error)
	Owns(context.Context, string, string) error
}

type Handler struct {
	sessions *session.Service
	store    Store
	signer   Signer
}

func NewHandler(sessions *session.Service, store Store, signer Signer) *Handler {
	return &Handler{sessions: sessions, store: store, signer: signer}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/documents", session.Require(h.sessions.Authenticate, h.registerDocument))
	mux.HandleFunc("GET /api/documents/{fileId}/content", session.Require(h.sessions.Authenticate, h.readContent))
	mux.HandleFunc("POST /api/documents/{fileId}/preview-links", session.Require(h.sessions.Authenticate, h.issuePreviewLink))
	mux.HandleFunc("GET /api/previews/{fileId}", h.servePreview)
}

func (h *Handler) registerDocument(w http.ResponseWriter, r *http.Request, owner string) {
	var input struct {
		Path string `json:"path"`
	}
	if middleware.Decode(w, r, &input) != nil {
		middleware.RespondError(w, fs.ErrInvalid)
		return
	}
	id, err := h.store.Register(r.Context(), owner, input.Path)
	middleware.Result(w, map[string]string{"fileId": id}, err)
}

func (h *Handler) readContent(w http.ResponseWriter, r *http.Request, owner string) {
	id := r.PathValue("fileId")
	if uuid.Validate(id) != nil {
		middleware.RespondError(w, fs.ErrInvalid)
		return
	}
	v, err := h.store.Document(r.Context(), owner, id)
	middleware.Content(w, "application/pdf", v, err)
}

func (h *Handler) issuePreviewLink(w http.ResponseWriter, r *http.Request, owner string) {
	id := r.PathValue("fileId")
	if uuid.Validate(id) != nil {
		middleware.RespondError(w, fs.ErrInvalid)
		return
	}
	if err := h.store.Owns(r.Context(), owner, id); err != nil {
		middleware.RespondError(w, err)
		return
	}
	middleware.WriteJSON(w, 200, h.signer.Issue(owner, id))
}

func (h *Handler) servePreview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	id := r.PathValue("fileId")
	owner, err := h.signer.Redeem(id, r.Method, r.URL.Query().Get("token"))
	if err != nil {
		middleware.Reject(w, 403, "forbidden")
		return
	}
	data, err := h.store.Document(r.Context(), owner, id)
	middleware.Content(w, "application/pdf", data, err)
}

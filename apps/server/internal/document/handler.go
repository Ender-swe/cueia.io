package document

import (
	"errors"
	"io/fs"
	"net/http"

	"github.com/Ender-swe/cueia.io/apps/server/internal/middleware"
)

type Handler struct {
	documents *Service
}

func NewHandler(store Store, signer Signer) *Handler {
	return &Handler{documents: NewService(store, signer)}
}

func (h *Handler) RegisterDocument(w http.ResponseWriter, r *http.Request, owner string) {
	var input struct {
		Path string `json:"path"`
	}
	if middleware.Decode(w, r, &input) != nil {
		middleware.RespondError(w, fs.ErrInvalid)
		return
	}
	id, err := h.documents.Register(r.Context(), owner, input.Path)
	middleware.Result(w, map[string]string{"fileId": id}, err)
}

func (h *Handler) ReadContent(w http.ResponseWriter, r *http.Request, owner string) {
	data, err := h.documents.Read(r.Context(), owner, r.PathValue("fileId"))
	middleware.Content(w, "application/pdf", data, err)
}

func (h *Handler) IssuePreviewLink(w http.ResponseWriter, r *http.Request, owner string) {
	link, err := h.documents.IssuePreviewLink(r.Context(), owner, r.PathValue("fileId"))
	middleware.Result(w, link, err)
}

func (h *Handler) ServePreview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	id := r.PathValue("fileId")
	data, err := h.documents.OpenPreview(r.Context(), id, r.Method, r.URL.Query().Get("token"))
	if errors.Is(err, ErrInvalidPreview) {
		middleware.Reject(w, 403, "forbidden")
		return
	}
	middleware.Content(w, "application/pdf", data, err)
}

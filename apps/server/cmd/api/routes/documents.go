package routes

import (
	"context"
	"net/http"

	"github.com/Ender-swe/cueia.io/apps/server/internal/document"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
)

func RegisterDocuments(mux *http.ServeMux, authenticate func(context.Context, string) (string, error), h *document.Handler) {
	mux.HandleFunc("POST /documents", session.Require(authenticate, h.RegisterDocument))
	mux.HandleFunc("GET /documents/{fileId}/content", session.Require(authenticate, h.ReadContent))
	mux.HandleFunc("POST /documents/{fileId}/preview-links", session.Require(authenticate, h.IssuePreviewLink))
	mux.HandleFunc("GET /previews/{fileId}", h.ServePreview)
}

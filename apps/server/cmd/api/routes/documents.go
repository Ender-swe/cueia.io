package routes

import (
	"context"
	"net/http"

	"github.com/Ender-swe/cueia.io/apps/server/internal/document"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
)

func registerDocuments(mux *http.ServeMux, authenticate func(context.Context, string) (string, error), h *document.Handler) {
	mux.HandleFunc("POST /api/documents", session.Require(authenticate, h.RegisterDocument))
	mux.HandleFunc("GET /api/documents/{fileId}/content", session.Require(authenticate, h.ReadContent))
	mux.HandleFunc("POST /api/documents/{fileId}/preview-links", session.Require(authenticate, h.IssuePreviewLink))
	mux.HandleFunc("GET /api/previews/{fileId}", h.ServePreview)
}

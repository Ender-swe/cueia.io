package main

import (
	"io/fs"
	"net/http"

	"github.com/Ender-swe/cueia.io/apps/server/internal/document"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace"
	"github.com/google/uuid"
)

func previewRoutes(mux *http.ServeMux, s session.Service, store *workspace.Service, signer document.Signer) {
	mux.HandleFunc("POST /api/documents/{fileId}/preview-links", requireSession(s.Authenticate, func(w http.ResponseWriter, r *http.Request, owner string) {
		id := r.PathValue("fileId")
		if uuid.Validate(id) != nil {
			respondError(w, fs.ErrInvalid)
			return
		}
		if err := store.Owns(r.Context(), owner, id); err != nil {
			respondError(w, err)
			return
		}
		writeJSON(w, 200, signer.Issue(owner, id))
	}))
	mux.HandleFunc("GET /api/previews/{fileId}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private, no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		id := r.PathValue("fileId")
		owner, err := signer.Redeem(id, r.Method, r.URL.Query().Get("token"))
		if err != nil {
			writeError(w, 403, "forbidden")
			return
		}
		data, err := store.Document(r.Context(), owner, id)
		content(w, "application/pdf", data, err)
	})
}

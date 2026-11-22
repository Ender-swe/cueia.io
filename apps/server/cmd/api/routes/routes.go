package routes

import (
	"context"
	"net/http"
	"time"

	"github.com/Ender-swe/cueia.io/apps/server/internal/document"
	"github.com/Ender-swe/cueia.io/apps/server/internal/middleware"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace"
)

type OptionalProvider struct {
	Name string
	// A nil check means the integration has no configuration.
	Check func(context.Context) error
}

type providerStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// New builds the API handler with session protection and rate limiting.
func New(sessions *session.Service, store *workspace.Service, signer document.Signer) http.Handler {
	mux := http.NewServeMux()
	RegisterStatus(mux, sessions.Authenticate, nil)
	RegisterSessions(mux, sessions)
	RegisterWorkspace(mux, sessions.Authenticate, workspace.NewHandler(store))
	RegisterDocuments(mux, sessions.Authenticate, document.NewHandler(store, signer))
	const prefix = "/api"
	root := http.NewServeMux()
	root.Handle(prefix+"/", http.StripPrefix(prefix, mux))
	return sessions.Protect(root, middleware.NewLimiter(), prefix+sessionPath)
}

func RegisterStatus(mux *http.ServeMux, authenticate func(context.Context, string) (string, error), providers []OptionalProvider) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /providers", session.Require(authenticate, func(w http.ResponseWriter, r *http.Request, _ string) {
		statuses := make([]providerStatus, 0, len(providers))
		for _, p := range providers {
			status := "disabled"
			if p.Check != nil {
				ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
				err := p.Check(ctx)
				cancel()
				status = "configured"
				if err != nil {
					status = "unavailable"
				}
			}
			statuses = append(statuses, providerStatus{p.Name, status})
		}
		middleware.WriteJSON(w, http.StatusOK, statuses)
	}))
}

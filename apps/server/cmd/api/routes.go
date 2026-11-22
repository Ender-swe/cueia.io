package main

import (
	"context"
	"net/http"
	"time"

	"github.com/Ender-swe/cueia.io/apps/server/internal/middleware"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
)

type optionalProvider struct {
	name string
	// A nil check means the integration has no configuration.
	check func(context.Context) error
}

type providerStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func routes(authenticate func(context.Context, string) (string, error), providers []optionalProvider) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		middleware.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/providers", session.Require(authenticate, func(w http.ResponseWriter, r *http.Request, _ string) {
		statuses := make([]providerStatus, 0, len(providers))
		for _, p := range providers {
			status := "disabled"
			if p.check != nil {
				ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
				err := p.check(ctx)
				cancel()
				status = "configured"
				if err != nil {
					status = "unavailable"
				}
			}
			statuses = append(statuses, providerStatus{p.name, status})
		}
		middleware.WriteJSON(w, http.StatusOK, statuses)
	}))
	return mux
}

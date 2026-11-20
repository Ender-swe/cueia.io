package main

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"time"

	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
	"github.com/jackc/pgx/v5"
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
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/providers", requireSession(authenticate, func(w http.ResponseWriter, r *http.Request, _ string) {
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
		writeJSON(w, http.StatusOK, statuses)
	}))
	return mux
}

type ownerHandler func(http.ResponseWriter, *http.Request, string)

func requireSession(authenticate func(context.Context, string) (string, error), next ownerHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		token := session.Bearer(r)
		if token == "" {
			writeError(w, 401, "unauthorized")
			return
		}
		owner, err := authenticate(r.Context(), token)
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, session.ErrUnauthorized) {
			writeError(w, 401, "unauthorized")
			return
		}
		if err != nil {
			writeError(w, 503, "unavailable")
			return
		}
		next(w, r, owner)
	}
}

func respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, fs.ErrInvalid):
		writeError(w, 400, "invalid_request")
	case errors.Is(err, fs.ErrNotExist), errors.Is(err, pgx.ErrNoRows):
		writeError(w, 404, "not_found")
	default:
		writeError(w, 503, "unavailable")
	}
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"code": code, "message": http.StatusText(status)})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

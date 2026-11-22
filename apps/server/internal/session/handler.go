package session

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"net"
	"net/http"

	"github.com/Ender-swe/cueia.io/apps/server/internal/middleware"
	"github.com/jackc/pgx/v5"
)

type Handler struct {
	sessions *Service
}

func NewHandler(sessions *Service) *Handler {
	return &Handler{sessions: sessions}
}

func (h *Handler) ExchangeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	credential, err := h.sessions.Exchange(r.Context(), Bearer(r))
	if err != nil {
		status, code := http.StatusServiceUnavailable, "unavailable"
		if errors.Is(err, ErrUnauthorized) {
			status, code = http.StatusUnauthorized, "unauthorized"
		}
		reject(w, status, code)
		return
	}
	_ = json.NewEncoder(w).Encode(credential)
}

func (h *Handler) RevokeHTTP(w http.ResponseWriter, r *http.Request, _ string) {
	if err := h.sessions.Revoke(r.Context(), Bearer(r)); err != nil {
		switch {
		case errors.Is(err, fs.ErrInvalid):
			reject(w, 400, "invalid_request")
		case errors.Is(err, fs.ErrNotExist), errors.Is(err, pgx.ErrNoRows):
			reject(w, 404, "not_found")
		default:
			reject(w, 503, "unavailable")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type OwnerHandler func(http.ResponseWriter, *http.Request, string)

func Require(authenticate func(context.Context, string) (string, error), next OwnerHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		token := Bearer(r)
		if token == "" {
			reject(w, 401, "unauthorized")
			return
		}
		owner, err := authenticate(r.Context(), token)
		if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, ErrUnauthorized) {
			reject(w, 401, "unauthorized")
			return
		}
		if err != nil {
			reject(w, 503, "unavailable")
			return
		}
		next(w, r, owner)
	}
}

func reject(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"code": code, "message": http.StatusText(status)})
}

// Count an authentication failure only after the handler has checked the credential.
type failureWriter struct {
	http.ResponseWriter
	count     func() bool
	throttled bool
}

func (w *failureWriter) WriteHeader(status int) {
	if status == http.StatusUnauthorized && !w.count() {
		w.throttled = true
		reject(w.ResponseWriter, 429, "rate_limited")
		return
	}
	w.ResponseWriter.WriteHeader(status)
}
func (w *failureWriter) Write(data []byte) (int, error) {
	if w.throttled {
		return len(data), nil
	}
	return w.ResponseWriter.Write(data)
}

func Protect(next http.Handler, origin string, limiter *middleware.Limiter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		origins := r.Header.Values("Origin")
		if len(origins) > 0 {
			if len(origins) != 1 || origins[0] != origin {
				reject(w, 403, "forbidden")
				return
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.WriteHeader(204)
				return
			}
		}
		peer, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			peer = r.RemoteAddr
		}
		if r.Method == "POST" && r.URL.Path == "/api/session" && limiter.Exhausted(peer) {
			reject(w, 429, "rate_limited")
			return
		}
		next.ServeHTTP(&failureWriter{ResponseWriter: w, count: func() bool { return limiter.Allow(peer) }}, r)
	})
}

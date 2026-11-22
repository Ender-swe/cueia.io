package session

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
)

type Handler struct {
	sessions *Service
}

func NewHandler(sessions *Service) *Handler {
	return &Handler{sessions: sessions}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc(exchangeRoute, h.ExchangeHTTP)
	mux.HandleFunc(revokeRoute, Require(h.sessions.Authenticate, h.RevokeHTTP))
}

func (h *Handler) ExchangeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	credential, err := h.sessions.Exchange(r.Context(), bearerToken(r))
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
	if err := h.sessions.Revoke(r.Context(), bearerToken(r)); err != nil {
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
		token := bearerToken(r)
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

// bearerToken reads the single credential carried in the Authorization header.
func bearerToken(r *http.Request) string {
	headers := r.Header.Values("Authorization")
	if len(headers) != 1 {
		return ""
	}
	fields := strings.Fields(headers[0])
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
		return ""
	}
	return fields[1]
}

func reject(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"code": code, "message": http.StatusText(status)})
}

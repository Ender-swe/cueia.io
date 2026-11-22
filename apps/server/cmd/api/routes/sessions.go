package routes

import (
	"net/http"

	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
)

func RegisterSessions(mux *http.ServeMux, service *session.Service) {
	h := session.NewHandler(service)
	mux.HandleFunc("POST /session", h.ExchangeHTTP)
	mux.HandleFunc("DELETE /session", session.Require(service.Authenticate, h.RevokeHTTP))
}

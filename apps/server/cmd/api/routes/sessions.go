package routes

import (
	"net/http"

	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
)

const sessionPath = "/session"

func RegisterSessions(mux *http.ServeMux, service *session.Service) {
	h := session.NewHandler(service)
	mux.HandleFunc("POST "+sessionPath, h.ExchangeHTTP)
	mux.HandleFunc("DELETE "+sessionPath, session.Require(service.Authenticate, h.RevokeHTTP))
}

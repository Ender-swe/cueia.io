package session

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// Bearer reads the single credential carried in the Authorization header.
func Bearer(r *http.Request) string {
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

func (s *Service) ExchangeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json")
	credential, err := s.Exchange(r.Context(), Bearer(r))
	if err != nil {
		status, code := http.StatusServiceUnavailable, "unavailable"
		if errors.Is(err, ErrUnauthorized) {
			status, code = http.StatusUnauthorized, "unauthorized"
		}
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]string{"code": code, "message": http.StatusText(status)})
		return
	}
	_ = json.NewEncoder(w).Encode(credential)
}

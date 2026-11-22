package session

import (
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

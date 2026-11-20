package middleware

import (
	"encoding/json"
	"net"
	"net/http"
)

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

func Protect(next http.Handler, origin string, limiter *Limiter) http.Handler {
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
		if r.Method == "POST" && r.URL.Path == "/api/session" && limiter.exhausted(peer) {
			reject(w, 429, "rate_limited")
			return
		}
		next.ServeHTTP(&failureWriter{ResponseWriter: w, count: func() bool { return limiter.Allow(peer) }}, r)
	})
}

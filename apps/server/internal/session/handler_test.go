package session

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ender-swe/cueia.io/apps/server/internal/middleware"
	"github.com/Ender-swe/cueia.io/apps/server/internal/testutil"
)

func TestOriginAndRateLimit(t *testing.T) {
	now := time.Now()
	l := middleware.NewLimiter()
	l.Limit = 2
	l.Now = func() time.Time { return now }
	calls := 0
	h := Protect(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") == "Bearer valid" {
			w.WriteHeader(200)
			return
		}
		reject(w, 401, "unauthorized")
	}), "https://dashboard.example", l)
	try := func(peer, origin, auth string, want int) {
		t.Helper()
		r := httptest.NewRequest("POST", "/api/session", nil)
		r.RemoteAddr = peer
		r.Header.Set("Origin", origin)
		r.Header.Set("Authorization", auth)
		r.Header.Set("X-Forwarded-For", peer)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		if rec.Code != want {
			t.Fatalf("got %d want %d", rec.Code, want)
		}
	}
	try("127.0.0.1:1", "https://other.example", "Bearer valid", 403)
	if calls != 0 {
		t.Fatal("origin check reached handler")
	}
	for i, want := range []int{401, 401, 429} {
		try([]string{"127.0.0.1:1", "127.0.0.1:2", "127.0.0.1:3"}[i], "https://dashboard.example", "", want)
	}
	try("127.0.0.2:1", "https://dashboard.example", "", 401)
	now = now.Add(l.Duration)
	try("127.0.0.1:4", "https://dashboard.example", "", 401)
}

func TestSessionRoutes(t *testing.T) {
	store := testutil.Workspace(t, "sample")
	service := NewService(store.DB, func(_ context.Context, token string) (string, error) {
		if token != "clerk-fixture" {
			return "", errors.New("invalid credential")
		}
		return "user", nil
	}, store.Provision)
	mux := http.NewServeMux()
	NewHandler(service).Register(mux)
	call := func(method, token string, status int) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, "/api/session", nil)
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != status {
			t.Fatalf("%s /api/session: got %d, want %d", method, w.Code, status)
		}
		return w
	}
	call("POST", "", 401)
	call("POST", "malformed", 401)
	res := call("POST", "clerk-fixture", 200)
	var body struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expiresAt"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil || body.Token == "" || !body.ExpiresAt.After(time.Now()) {
		t.Fatalf("invalid exchange response: %v", err)
	}
	if res := call("DELETE", body.Token, 204); res.Body.Len() != 0 {
		t.Error("unexpected revocation body")
	}
	call("DELETE", body.Token, 401)
}

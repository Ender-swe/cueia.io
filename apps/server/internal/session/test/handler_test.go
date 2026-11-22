package sessiontest_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ender-swe/cueia.io/apps/server/internal/middleware"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session/test"
)

func TestOriginAndRateLimit(t *testing.T) {
	now := time.Now()
	l := middleware.NewLimiter()
	l.Limit = 2
	l.Now = func() time.Time { return now }
	calls := 0
	t.Setenv("CLERK_ISSUER", "https://issuer.example")
	t.Setenv("CLERK_AUTHORIZED_PARTY", sessiontest.Origin)
	service, err := session.NewService(t.Context(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	h := service.Protect(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") == "Bearer valid" {
			w.WriteHeader(200)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}), l, "/api/session")
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

package routes_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ender-swe/cueia.io/apps/server/cmd/api/routes"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session/test"
)

func TestSessionRoutes(t *testing.T) {
	store := testWorkspace(t, "sample")
	issuer := sessiontest.NewIssuer(t)
	t.Setenv("CLERK_ISSUER", issuer.URL)
	t.Setenv("CLERK_AUTHORIZED_PARTY", sessiontest.Origin)
	service, err := session.NewService(t.Context(), store.DB, store.Provision)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	routes.RegisterSessions(mux, service)
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
	res := call("POST", issuer.Token(t, "user"), 200)
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

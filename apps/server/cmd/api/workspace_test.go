package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Ender-swe/cueia.io/apps/server/internal/database"
	"github.com/Ender-swe/cueia.io/apps/server/internal/document"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session/test"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider/providertest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func request(h http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func login(t *testing.T, handler http.Handler, clerkToken string) session.Credential {
	t.Helper()
	rec := request(handler, "POST", "/api/session", clerkToken, "")
	var c session.Credential
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &c) != nil || c.Token == "" || !c.ExpiresAt.After(time.Now()) {
		t.Fatalf("session exchange failed (status %d)", rec.Code)
	}
	return c
}

func TestWorkspaceHTTP(t *testing.T) {
	store := testWorkspace(t, "sample")
	issuer := sessiontest.NewIssuer(t)
	t.Setenv("CLERK_ISSUER", issuer.URL)
	t.Setenv("CLERK_AUTHORIZED_PARTY", sessiontest.Origin)
	s, err := session.NewService(t.Context(), store.DB, store.Provision)
	if err != nil {
		t.Fatal(err)
	}
	mux := routes(s.Authenticate, nil)
	workspaceRoutes(mux, s, store)
	signer, err := document.Load(t.Context(), store.DB)
	if err != nil {
		t.Fatal(err)
	}
	document.NewHandler(s, store, signer).Register(mux)
	alice := uuid.NewString()
	a := login(t, mux, issuer.Token(t, alice)).Token
	b := login(t, mux, issuer.Token(t, alice)).Token
	bob := login(t, mux, issuer.Token(t, uuid.NewString())).Token
	check := func(method, path, token, body string, status int) []byte {
		t.Helper()
		rec := request(mux, method, path, token, body)
		if rec.Code != status {
			t.Fatalf("%s %s: got %d want %d", method, path, rec.Code, status)
		}
		return rec.Body.Bytes()
	}
	check("GET", "/api/workspace", "", "", 401)
	check("PATCH", "/api/workspace/preferences", a, `{"timeFormat":"24h"}`, 200)
	var snap workspace.Snapshot
	if json.Unmarshal(check("GET", "/api/workspace", b, "", 200), &snap) != nil || snap.Preferences["timeFormat"] != "24h" {
		t.Fatal("preference not shared")
	}
	if json.Unmarshal(check("GET", "/api/workspace", bob, "", 200), &snap) != nil || snap.Preferences["timeFormat"] != "12h" {
		t.Fatal("owners shared preferences")
	}
	for _, path := range []string{"/api/mail/threads", "/api/calendar/events"} {
		if string(check("GET", path, a, "", 200)) == string(check("GET", path, bob, "", 200)) {
			t.Fatal("owners shared collection", path)
		}
	}
	check("GET", "/api/files", a, "", 200)
	check("GET", "/api/files/search?name=registration", a, "", 200)
	file := "Attachments/north-pier-registration.pdf"
	raw := check("GET", "/api/files/content?path="+file, a, "", 200)
	var doc struct {
		FileID string `json:"fileId"`
	}
	if json.Unmarshal(check("POST", "/api/documents", a, `{"path":"`+file+`"}`, 200), &doc) != nil || doc.FileID == "" {
		t.Fatal("registration failed")
	}
	path := "/api/documents/" + doc.FileID + "/content"
	if string(check("GET", path, a, "", 200)) != string(raw) {
		t.Fatal("wrong document content")
	}
	check("GET", path, bob, "", 404)
	for _, tc := range []struct {
		method, path, body string
		status             int
		code, message      string
	}{
		{"PATCH", "/api/workspace/preferences", `{"timeFormat":null}`, 400, "invalid_request", "Bad Request"},
		{"GET", "/api/documents/" + uuid.NewString() + "/content", "", 404, "not_found", "Not Found"},
	} {
		var e struct{ Code, Message string }
		data := check(tc.method, tc.path, a, tc.body, tc.status)
		if json.Unmarshal(data, &e) != nil || e.Code != tc.code || e.Message != tc.message {
			t.Fatal("unexpected error body", tc.status)
		}
	}
	fixture := store.Provider.(*providertest.Fixture)
	dir := fixture.Directory
	fixture.Directory = t.TempDir()
	data := check("GET", "/api/files", a, "", 503)
	if strings.TrimSpace(string(data)) != `{"code":"unavailable","message":"Service Unavailable"}` {
		t.Fatal("unexpected unavailable response")
	}
	fixture.Directory = dir
}

func testWorkspace(t *testing.T, mode string) *workspace.Service {
	t.Helper()
	dsn := os.Getenv("CUEIA_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires CUEIA_TEST_DATABASE_URL")
	}
	db, err := pgxpool.New(t.Context(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	schema := "api_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = db.Exec(t.Context(), "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	admin := db
	t.Cleanup(func() {
		if _, err := admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Error(err)
		}
	})
	config := db.Config()
	config.ConnConfig.RuntimeParams["search_path"] = schema
	db, err = pgxpool.NewWithConfig(t.Context(), config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	conn, err := db.Acquire(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	err = database.Migrate(t.Context(), conn.Conn(), os.DirFS("../../migrations"))
	conn.Release()
	if err != nil {
		t.Fatal(err)
	}
	return &workspace.Service{DB: db, Provider: &providertest.Fixture{Directory: t.TempDir()}, Deployment: uuid.NewString(), ProviderName: "internal", Region: "local", Mode: mode}
}

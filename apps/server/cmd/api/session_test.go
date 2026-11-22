package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"github.com/Ender-swe/cueia.io/apps/server/internal/document"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/sample"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Ender-swe/cueia.io/apps/server/internal/database"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
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

func login(t *testing.T, handler http.Handler, subject string) session.Credential {
	t.Helper()
	rec := request(handler, "POST", "/api/session", subject, "")
	var c session.Credential
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &c) != nil || c.Token == "" || !c.ExpiresAt.After(time.Now()) {
		t.Fatalf("session exchange failed (status %d)", rec.Code)
	}
	return c
}

func TestSessionExchange(t *testing.T) {
	for _, mode := range []string{"sample", "live"} {
		t.Run(mode, func(t *testing.T) {
			store := testWorkspace(t, mode)
			s := session.NewService(store.DB, session.Clerk(t.Context(), "https://issuer.invalid", "https://cueia.invalid"), store.Provision)
			mux := http.NewServeMux()
			mux.HandleFunc("POST /api/session", s.ExchangeHTTP)
			for _, token := range []string{"", "malformed"} {
				if request(mux, "POST", "/api/session", token, "").Code != 401 {
					t.Fatal("invalid credential accepted")
				}
			}
			sub := uuid.NewString()
			s = session.NewService(store.DB, func(context.Context, string) (string, error) { return sub, nil }, store.Provision)
			handler := http.HandlerFunc(s.ExchangeHTTP)
			a, b := login(t, handler, "fixture"), login(t, handler, "fixture")
			if a.Token == b.Token {
				t.Fatal("repeated session credential")
			}
			for _, c := range []session.Credential{a, b} {
				if _, err := s.Authenticate(t.Context(), c.Token); err != nil {
					t.Fatal(err)
				}
				digest := sha256.Sum256([]byte(c.Token))
				var expires time.Time
				if err := store.DB.QueryRow(t.Context(), "SELECT expires_at FROM sessions WHERE token_digest=$1", digest[:]).Scan(&expires); err != nil || !expires.Equal(c.ExpiresAt) {
					t.Fatal("persisted credential mismatch", err)
				}
			}
		})
	}
}

func TestWorkspaceHTTP(t *testing.T) {
	store := testWorkspace(t, "sample")
	s := session.NewService(store.DB, func(_ context.Context, subject string) (string, error) { return subject, nil }, store.Provision)
	mux := routes(s.Authenticate, nil)
	workspaceRoutes(mux, s, store)
	alice := uuid.NewString()
	a := login(t, mux, alice).Token
	b := login(t, mux, alice).Token
	bob := login(t, mux, uuid.NewString()).Token
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
	check("DELETE", "/api/session", a, "", 204)
	check("GET", "/api/workspace", a, "", 401)
	check("GET", "/api/workspace", b, "", 200)
	digest := sha256.Sum256([]byte(b))
	if _, err := store.DB.Exec(t.Context(), "UPDATE sessions SET expires_at=now()-interval '1 second' WHERE token_digest=$1", digest[:]); err != nil {
		t.Fatal(err)
	}
	check("GET", "/api/workspace", b, "", 401)
	check("GET", "/api/workspace", bob, "", 200)
}

func TestPreviewAcrossServers(t *testing.T) {
	store := testWorkspace(t, "sample")
	s := session.NewService(store.DB, func(_ context.Context, sub string) (string, error) { return sub, nil }, store.Provision)
	var servers [2]*httptest.Server
	var errs [2]error
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range servers {
		wg.Go(func() {
			<-start
			signer, err := document.Load(t.Context(), store.DB)
			if err != nil {
				errs[i] = err
				return
			}
			mux := routes(s.Authenticate, nil)
			workspaceRoutes(mux, s, store)
			previewRoutes(mux, s, store, signer)
			servers[i] = httptest.NewServer(mux)
		})
	}
	close(start)
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(servers[i].Close)
	}
	a := login(t, servers[0].Config.Handler, uuid.NewString()).Token
	b := login(t, servers[1].Config.Handler, uuid.NewString()).Token
	rec := request(servers[0].Config.Handler, "POST", "/api/documents", a, `{"path":"Attachments/north-pier-registration.pdf"}`)
	var doc struct{ FileID string }
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &doc) != nil {
		t.Fatal("register", rec.Code)
	}
	path := "/api/documents/" + doc.FileID + "/preview-links"
	if request(servers[0].Config.Handler, "POST", path, b, "").Code != 404 {
		t.Fatal("foreign preview issued")
	}
	rec = request(servers[0].Config.Handler, "POST", path, a, "")
	var link document.Link
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &link) != nil {
		t.Fatal("issue", rec.Code)
	}
	res, err := servers[1].Client().Get(servers[1].URL + link.URL)
	if err != nil {
		t.Fatal("redemption request failed")
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil || res.StatusCode != 200 || !bytes.Equal(data, sample.PDF) {
		t.Fatal("cross-instance preview mismatch")
	}
	u, _ := url.Parse(link.URL)
	for _, tc := range []struct{ method, path string }{{"GET", "/api/files"}, {"POST", "/api/documents"}} {
		if request(servers[1].Config.Handler, tc.method, tc.path, u.Query().Get("token"), "").Code != 401 {
			t.Fatal("preview granted workspace access")
		}
	}
}

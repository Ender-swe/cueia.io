package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
	"github.com/Ender-swe/cueia.io/apps/server/internal/testutil"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/provider/providertest"
	"github.com/google/uuid"
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

func login(t *testing.T, handler http.Handler, subject string) session.Credential {
	t.Helper()
	rec := request(handler, "POST", "/api/session", subject, "")
	var c session.Credential
	if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &c) != nil || c.Token == "" || !c.ExpiresAt.After(time.Now()) {
		t.Fatalf("session exchange failed (status %d)", rec.Code)
	}
	return c
}

func TestWorkspaceHTTP(t *testing.T) {
	store := testutil.Workspace(t, "sample")
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
}

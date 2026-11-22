package main

import (
	"bytes"

	"encoding/json"
	"io"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/Ender-swe/cueia.io/apps/server/internal/document"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session/sessiontest"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/sample"
	"github.com/google/uuid"
)

func TestPreviewAcrossServers(t *testing.T) {
	store := testWorkspace(t, "sample")
	issuer := sessiontest.NewIssuer(t)
	t.Setenv("CLERK_ISSUER", issuer.URL)
	t.Setenv("CLERK_AUTHORIZED_PARTY", sessiontest.Origin)
	s, err := session.NewService(t.Context(), store.DB, store.Provision)
	if err != nil {
		t.Fatal(err)
	}
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
	a := login(t, servers[0].Config.Handler, issuer.Token(t, uuid.NewString())).Token
	b := login(t, servers[1].Config.Handler, issuer.Token(t, uuid.NewString())).Token
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

package document_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ender-swe/cueia.io/apps/server/internal/document"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/sample"
)

func TestPreviewRoute(t *testing.T) {
	signer := document.Signer{Now: time.Now}
	mux := http.NewServeMux()
	document.NewHandler(nil, fileStore{sample.PDF}, signer).Register(mux)
	link := signer.Issue("owner", fileID)
	for _, tc := range []struct {
		name, url string
		status    int
	}{
		{"signed link", link.URL, http.StatusOK},
		{"missing token", "/api/previews/" + fileID, http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.url, nil))
			if rec.Code != tc.status {
				t.Fatalf("got HTTP %d, want %d", rec.Code, tc.status)
			}
			if got := rec.Header().Get("Cache-Control"); got != "private, no-store" {
				t.Errorf("unexpected Cache-Control %q", got)
			}
			if got := rec.Header().Get("Referrer-Policy"); got != "no-referrer" {
				t.Errorf("unexpected Referrer-Policy %q", got)
			}
			if tc.status == http.StatusOK {
				if got := rec.Header().Get("Content-Type"); got != "application/pdf" {
					t.Errorf("unexpected Content-Type %q", got)
				}
				if !bytes.Equal(rec.Body.Bytes(), sample.PDF) {
					t.Error("unexpected PDF response body")
				}
			}
		})
	}
}

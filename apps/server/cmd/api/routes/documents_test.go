package routes

import (
	"bytes"
	"context"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ender-swe/cueia.io/apps/server/internal/document"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/sample"
)

type previewStore struct {
	document.Store
}

const fileID = "12345678-1234-4123-8123-123456789abc"

func (previewStore) ReadFileVersion(_ context.Context, owner, id string) ([]byte, error) {
	if owner != "owner" || id != fileID {
		return nil, fs.ErrNotExist
	}
	return sample.PDF, nil
}

func TestPreviewRoute(t *testing.T) {
	signer := document.Signer{Now: time.Now}
	mux := http.NewServeMux()
	registerDocuments(mux, nil, document.NewHandler(previewStore{}, signer))
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

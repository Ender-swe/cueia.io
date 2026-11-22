package document_test

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"testing"
	"time"

	"github.com/Ender-swe/cueia.io/apps/server/internal/document"
	"github.com/Ender-swe/cueia.io/apps/server/internal/workspace/sample"
)

const fileID = "12345678-1234-4123-8123-123456789abc"

type fileStore struct {
	data []byte
}

func (s fileStore) RegisterFileVersion(_ context.Context, owner, name, mediaType string, validate func([]byte) error) (string, error) {
	if owner != "owner" || name != "agenda.PDF" || mediaType != "application/pdf" {
		return "", errors.New("unexpected registration parameters")
	}
	if err := validate(s.data); err != nil {
		return "", err
	}
	return fileID, nil
}

func (s fileStore) Owns(_ context.Context, owner, id string) error {
	if owner != "owner" || id != fileID {
		return fs.ErrNotExist
	}
	return nil
}

func (s fileStore) ReadFileVersion(ctx context.Context, owner, id string) ([]byte, error) {
	if err := s.Owns(ctx, owner, id); err != nil {
		return nil, err
	}
	return s.data, nil
}

func TestRegisterPDF(t *testing.T) {
	for _, tc := range []struct {
		name, path string
		data       []byte
		wantID     string
		wantErr    error
	}{
		{"uppercase extension", "agenda.PDF", sample.PDF, fileID, nil},
		{"wrong extension", "agenda.txt", sample.PDF, "", fs.ErrInvalid},
		{"missing PDF header", "agenda.PDF", []byte("plain text"), "", fs.ErrInvalid},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := document.NewService(fileStore{tc.data}, document.Signer{})
			id, err := service.Register(t.Context(), "owner", tc.path)
			if id != tc.wantID || !errors.Is(err, tc.wantErr) {
				t.Fatalf("got (%q, %v), want (%q, %v)", id, err, tc.wantID, tc.wantErr)
			}
		})
	}
}

func TestReadAndPreview(t *testing.T) {
	service := document.NewService(fileStore{sample.PDF}, document.Signer{Now: time.Now})
	ctx := t.Context()
	data, err := service.Read(ctx, "owner", fileID)
	if err != nil || !bytes.Equal(data, sample.PDF) {
		t.Fatalf("unexpected document content, error %v", err)
	}
	link, err := service.IssuePreviewLink(ctx, "owner", fileID)
	if err != nil {
		t.Fatal(err)
	}
	data, err = service.OpenPreview(ctx, fileID, "GET", tokenFrom(t, link))
	if err != nil || !bytes.Equal(data, sample.PDF) {
		t.Fatalf("unexpected preview content, error %v", err)
	}
	if _, err := service.IssuePreviewLink(ctx, "other", fileID); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected ownership failure, got %v", err)
	}
	if _, err := service.Read(ctx, "owner", "invalid"); !errors.Is(err, fs.ErrInvalid) {
		t.Fatalf("expected invalid file ID on read, got %v", err)
	}
	if _, err := service.IssuePreviewLink(ctx, "owner", "invalid"); !errors.Is(err, fs.ErrInvalid) {
		t.Fatalf("expected invalid file ID on preview, got %v", err)
	}
}

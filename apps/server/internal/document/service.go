package document

import (
	"bytes"
	"context"
	"io/fs"
	"path"
	"strings"

	"github.com/google/uuid"
)

type Store interface {
	RegisterFileVersion(ctx context.Context, owner, name, mediaType string, validate func([]byte) error) (string, error)
	ReadFileVersion(context.Context, string, string) ([]byte, error)
	Owns(context.Context, string, string) error
}

type Service struct {
	store  Store
	signer Signer
}

func NewService(store Store, signer Signer) *Service {
	return &Service{store: store, signer: signer}
}

func (s *Service) Register(ctx context.Context, owner, name string) (string, error) {
	if !strings.EqualFold(path.Ext(name), ".pdf") {
		return "", fs.ErrInvalid
	}
	return s.store.RegisterFileVersion(ctx, owner, name, "application/pdf", validatePDFHeader)
}

func validatePDFHeader(data []byte) error {
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		return fs.ErrInvalid
	}
	return nil
}

func (s *Service) Read(ctx context.Context, owner, fileID string) ([]byte, error) {
	if uuid.Validate(fileID) != nil {
		return nil, fs.ErrInvalid
	}
	return s.store.ReadFileVersion(ctx, owner, fileID)
}

func (s *Service) IssuePreviewLink(ctx context.Context, owner, fileID string) (Link, error) {
	if uuid.Validate(fileID) != nil {
		return Link{}, fs.ErrInvalid
	}
	if err := s.store.Owns(ctx, owner, fileID); err != nil {
		return Link{}, err
	}
	return s.signer.Issue(owner, fileID), nil
}

func (s *Service) OpenPreview(ctx context.Context, fileID, method, token string) ([]byte, error) {
	owner, err := s.signer.Redeem(fileID, method, token)
	if err != nil {
		return nil, err
	}
	return s.store.ReadFileVersion(ctx, owner, fileID)
}

package workspace

import (
	"context"
	"encoding/json"
)

func (s *Service) Snapshot(ctx context.Context, w string) (Snapshot, error) {
	return s.loadSnapshot(ctx, w)
}

func (s *Service) Preferences(ctx context.Context, w string, changes map[string]string) (map[string]string, error) {
	return s.updatePreferences(ctx, w, changes)
}

func (s *Service) Owns(ctx context.Context, w, id string) error {
	return s.checkFileOwner(ctx, w, id)
}

func (s *Service) Threads(ctx context.Context, w string) (json.RawMessage, error) {
	snapshot, err := s.Snapshot(ctx, w)
	return snapshot.Threads, err
}

func (s *Service) Events(ctx context.Context, w string) (json.RawMessage, error) {
	snapshot, err := s.Snapshot(ctx, w)
	return snapshot.Events, err
}

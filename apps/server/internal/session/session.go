package session

import (
	"context"
	"errors"
)

var ErrUnauthorized = errors.New("unauthorized")

// VerifyFunc returns the Clerk subject for a verified session token.
type VerifyFunc func(ctx context.Context, rawToken string) (subject string, err error)

// Verify checks the Clerk token and returns its subject.
func (s *Service) Verify(ctx context.Context, clerkToken string) (string, error) {
	if clerkToken == "" {
		return "", ErrUnauthorized
	}
	sub, err := s.verify(ctx, clerkToken)
	if err != nil {
		return "", ErrUnauthorized
	}
	return sub, nil
}

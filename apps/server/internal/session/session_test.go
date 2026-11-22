package session

import (
	"context"
	"errors"
	"testing"
)

func TestVerify(t *testing.T) {
	s := NewService(nil, func(_ context.Context, token string) (string, error) {
		switch token {
		case "":
			t.Fatal("empty token reached verifier")
		case "valid":
			return "user_123", nil
		}
		return "", errors.New("verification failed")
	}, nil)

	for _, token := range []string{"", "invalid"} {
		if sub, err := s.Verify(t.Context(), token); sub != "" || !errors.Is(err, ErrUnauthorized) {
			t.Errorf("Verify(%q) = (%q, %v), want empty subject and ErrUnauthorized", token, sub, err)
		}
	}
	if sub, err := s.Verify(t.Context(), "valid"); sub != "user_123" || err != nil {
		t.Errorf("Verify(valid) = (%q, %v), want user_123 and nil", sub, err)
	}
}

package document_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/Ender-swe/cueia.io/apps/server/internal/document"
)

func tokenFrom(t *testing.T, link document.Link) string {
	t.Helper()
	u, err := url.Parse(link.URL)
	if err != nil {
		t.Fatal(err)
	}
	token := u.Query().Get("token")
	if token == "" {
		t.Fatal("missing token in preview URL")
	}
	return token
}

func TestGrantBounds(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	s := document.Signer{Now: func() time.Time { return now }}
	link := s.Issue("owner", "document")
	if !link.ExpiresAt.Equal(now.Add(5 * time.Minute)) {
		t.Fatalf("unexpected expiry %v", link.ExpiresAt)
	}
	token := tokenFrom(t, link)
	if owner, err := s.Redeem("document", "GET", token); err != nil || owner != "owner" {
		t.Fatalf("redeemed owner %q, error %v", owner, err)
	}

	// Sign a different purpose with the test signer's empty key so the
	// signature is valid and the purpose check itself is exercised.
	raw, err := json.Marshal(map[string]any{
		"Workspace": "owner", "File": "document", "Purpose": "download",
		"Method": "GET", "Expiry": link.ExpiresAt.Unix(),
	})
	if err != nil {
		t.Fatal(err)
	}
	payload := base64.RawURLEncoding.EncodeToString(raw)
	mac := hmac.New(sha256.New, nil)
	mac.Write([]byte(payload))
	otherPurpose := payload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	for _, tc := range []struct{ name, id, method, token string }{
		{"wrong file", "other", "GET", token},
		{"wrong method", "document", "POST", token},
		{"tampering", "document", "GET", token + "X"},
		{"wrong purpose", "document", "GET", otherPurpose},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := s.Redeem(tc.id, tc.method, tc.token); !errors.Is(err, document.ErrInvalidPreview) {
				t.Fatalf("expected invalid preview, got %v", err)
			}
		})
	}
	now = link.ExpiresAt
	if _, err := s.Redeem("document", "GET", token); !errors.Is(err, document.ErrInvalidPreview) {
		t.Fatalf("expected invalid preview at expiry, got %v", err)
	}
}

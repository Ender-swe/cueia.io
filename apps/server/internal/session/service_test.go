package session

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestVerify(t *testing.T) {
	s := NewServiceWithVerifier(nil, func(_ context.Context, token string) (string, error) {
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

func TestClerkServiceVerify(t *testing.T) {
	newKey := func() *rsa.PrivateKey {
		t.Helper()
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatal(err)
		}
		return key
	}
	key := newKey()
	jwks := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, Use: "sig", Algorithm: "RS256"}}}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(jwks); err != nil {
			t.Error(err)
		}
	})
	issuer := httptest.NewServer(mux)
	defer issuer.Close()
	const origin = "https://dashboard.example"
	service := NewService(t.Context(), nil, Config{ClerkIssuer: issuer.URL + "/", AuthorizedParty: origin}, nil)
	for _, tc := range []struct {
		name, issuer, party, sid, sub string
		key                           *rsa.PrivateKey
		valid                         bool
	}{
		{"valid", issuer.URL, origin, "session", "user", key, true},
		{"invalid signature", issuer.URL, origin, "session", "user", newKey(), false},
		{"wrong issuer", "https://other.example", origin, "session", "user", key, false},
		{"wrong party", issuer.URL, "https://other.example", "session", "user", key, false},
		{"missing session ID", issuer.URL, origin, "", "user", key, false},
		{"missing subject", issuer.URL, origin, "session", "", key, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: tc.key}, nil)
			if err != nil {
				t.Fatal(err)
			}
			token, err := jwt.Signed(signer).Claims(jwt.Claims{Issuer: tc.issuer, Subject: tc.sub, Expiry: jwt.NewNumericDate(time.Now().Add(time.Hour))}).Claims(struct {
				Party   string `json:"azp"`
				Session string `json:"sid,omitempty"`
			}{tc.party, tc.sid}).Serialize()
			if err != nil {
				t.Fatal(err)
			}
			sub, err := service.Verify(t.Context(), token)
			if tc.valid {
				if sub != tc.sub || err != nil {
					t.Fatalf("valid token: subject %q, error %v", sub, err)
				}
			} else if sub != "" || !errors.Is(err, ErrUnauthorized) {
				t.Fatalf("invalid token: subject %q, error %v", sub, err)
			}
		})
	}
}

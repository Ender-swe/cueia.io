// Package sessiontest supplies a local signing-key endpoint for session tests.
package sessiontest

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

const Origin = "https://dashboard.example"

type Issuer struct {
	URL    string
	signer jose.Signer
}

func NewIssuer(t *testing.T) *Issuer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, nil)
	if err != nil {
		t.Fatal(err)
	}
	keys := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, Use: "sig", Algorithm: "RS256"}}}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(keys); err != nil {
			t.Error(err)
		}
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return &Issuer{URL: server.URL, signer: signer}
}

func (i *Issuer) Token(t *testing.T, subject string) string {
	t.Helper()
	return i.Sign(t, i.URL, Origin, "session", subject)
}

func (i *Issuer) Sign(t *testing.T, issuer, party, sid, subject string) string {
	t.Helper()
	claims := jwt.Claims{Issuer: issuer, Subject: subject, Expiry: jwt.NewNumericDate(time.Now().Add(time.Hour))}
	clerk := struct {
		Party   string `json:"azp"`
		Session string `json:"sid,omitempty"`
	}{party, sid}
	token, err := jwt.Signed(i.signer).Claims(claims).Claims(clerk).Serialize()
	if err != nil {
		t.Fatal(err)
	}
	return token
}

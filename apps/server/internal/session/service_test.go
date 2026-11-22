package session

import (
	"errors"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session/sessiontest"
	"testing"
)

func TestClerkServiceVerify(t *testing.T) {
	issuer := sessiontest.NewIssuer(t)
	const origin = sessiontest.Origin
	service := NewService(t.Context(), nil, Config{ClerkIssuer: issuer.URL + "/", AuthorizedParty: origin}, nil)
	for _, token := range []string{"", "invalid"} {
		if sub, err := service.Verify(t.Context(), token); sub != "" || !errors.Is(err, ErrUnauthorized) {
			t.Errorf("Verify(%q) = (%q, %v), want empty subject and ErrUnauthorized", token, sub, err)
		}
	}
	for _, tc := range []struct {
		name, issuer, party, sid, sub string
		signer                        *sessiontest.Issuer
		valid                         bool
	}{
		{"valid", issuer.URL, origin, "session", "user", issuer, true},
		{"invalid signature", issuer.URL, origin, "session", "user", sessiontest.NewIssuer(t), false},
		{"wrong issuer", "https://other.example", origin, "session", "user", issuer, false},
		{"wrong party", issuer.URL, "https://other.example", "session", "user", issuer, false},
		{"missing session ID", issuer.URL, origin, "", "user", issuer, false},
		{"missing subject", issuer.URL, origin, "session", "", issuer, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			token := tc.signer.Sign(t, tc.issuer, tc.party, tc.sid, tc.sub)
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

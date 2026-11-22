package sessiontest_test

import (
	"errors"
	"testing"

	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
	"github.com/Ender-swe/cueia.io/apps/server/internal/session/test"
)

func TestClerkServiceVerify(t *testing.T) {
	issuer := sessiontest.NewIssuer(t)
	const origin = sessiontest.Origin
	t.Setenv("CLERK_ISSUER", issuer.URL+"/")
	t.Setenv("CLERK_AUTHORIZED_PARTY", origin)
	service, err := session.NewService(t.Context(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, token := range []string{"", "invalid"} {
		if sub, err := service.Verify(t.Context(), token); sub != "" || !errors.Is(err, session.ErrUnauthorized) {
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
			} else if sub != "" || !errors.Is(err, session.ErrUnauthorized) {
				t.Fatalf("invalid token: subject %q, error %v", sub, err)
			}
		})
	}
}

package session

import (
	"slices"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	for _, tc := range []struct {
		name, issuer, party string
		missing             []string
	}{
		{"configured", "https://issuer.example", "https://cueia.example", nil},
		{"issuer missing", "", "https://cueia.example", []string{"CLERK_ISSUER"}},
		{"party missing", "https://issuer.example", "  ", []string{"CLERK_AUTHORIZED_PARTY"}},
		{"both missing", " ", "", []string{"CLERK_ISSUER", "CLERK_AUTHORIZED_PARTY"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CLERK_ISSUER", tc.issuer)
			t.Setenv("CLERK_AUTHORIZED_PARTY", tc.party)
			cfg, missing := loadConfig()
			if cfg.ClerkIssuer != tc.issuer || cfg.AuthorizedParty != tc.party || !slices.Equal(missing, tc.missing) {
				t.Fatalf("unexpected config %v, missing %v", cfg, missing)
			}
		})
	}
}

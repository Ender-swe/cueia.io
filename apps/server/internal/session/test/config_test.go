package sessiontest_test

import (
	"strings"
	"testing"

	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
)

func TestServiceConfiguration(t *testing.T) {
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
			_, err := session.NewService(t.Context(), nil, nil)
			if len(tc.missing) == 0 {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			want := "missing settings: " + strings.Join(tc.missing, ", ")
			if err == nil || err.Error() != want {
				t.Fatalf("got %v, want %q", err, want)
			}
		})
	}
}

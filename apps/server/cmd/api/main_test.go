package main

import (
	"strings"
	"testing"
)

func TestRequiredAPISettings(t *testing.T) {
	required := []string{"CLERK_ISSUER", "CLERK_AUTHORIZED_PARTY", "CPK_INTELLIGENCE_API_KEY"}
	for _, mode := range []string{"sample", "live"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("CUEIA_MODE", mode)
			t.Setenv("DATABASE_URL", "postgres://localhost/cueia")
			for _, name := range required {
				t.Setenv(name, "")
			}
			err := run(t.Context())
			if err == nil {
				t.Fatal("accepted empty API configuration")
			}
			for _, name := range required {
				if !strings.Contains(err.Error(), name) {
					t.Errorf("missing diagnostic for %s", name)
				}
			}
			const secret = "test-only-private-value"
			t.Setenv("CPK_INTELLIGENCE_API_KEY", secret)
			err = run(t.Context())
			if err == nil || strings.Contains(err.Error(), secret) {
				t.Fatal("expected a secret-free configuration error")
			}
		})
	}
}

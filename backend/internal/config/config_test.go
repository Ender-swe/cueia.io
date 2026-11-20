package config

import (
	"net"
	"strings"
	"testing"
)

func TestRequiredAPISettings(t *testing.T) {
	required := []string{"CLERK_ISSUER", "CLERK_AUTHORIZED_PARTY", "CPK_INTELLIGENCE_API_KEY"}
	for _, mode := range []string{"sample", "live"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("CUEIA_MODE", mode)
			for _, name := range required {
				t.Setenv(name, "")
			}
			_, err := LoadAPI()
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
			_, err = LoadAPI()
			if err == nil || strings.Contains(err.Error(), secret) {
				t.Fatal("expected a secret-free configuration error")
			}
		})
	}
}

func TestSampleDefaultListener(t *testing.T) {
	t.Setenv("CUEIA_MODE", "sample")
	t.Setenv("CUEIA_LISTEN_ADDR", "")
	t.Setenv("CLERK_ISSUER", "https://issuer.example")
	t.Setenv("CLERK_AUTHORIZED_PARTY", "http://localhost:5173")
	t.Setenv("CPK_INTELLIGENCE_API_KEY", "test-only-private-value")
	settings, err := LoadAPI()
	if err != nil {
		t.Fatal(err)
	}
	host, _, err := net.SplitHostPort(settings.Addr)
	if err != nil || !net.ParseIP(host).IsLoopback() {
		t.Fatal("default sample listener must bind to loopback")
	}
}

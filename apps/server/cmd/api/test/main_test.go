package api_test

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRequiredAPISettings(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "api")
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binary, "..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build API: %v\n%s", err, output)
	}
	startupError := func(t *testing.T) string {
		t.Helper()
		output, err := exec.CommandContext(t.Context(), binary).CombinedOutput()
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
			t.Fatalf("expected API to exit with status 1, got %v\n%s", err, output)
		}
		return string(output)
	}
	required := []string{"CLERK_ISSUER", "CLERK_AUTHORIZED_PARTY", "CPK_INTELLIGENCE_API_KEY"}
	for _, mode := range []string{"sample", "live"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("CUEIA_MODE", mode)
			t.Setenv("DATABASE_URL", "postgres://localhost/cueia")
			for _, name := range required {
				t.Setenv(name, "")
			}
			diagnostic := startupError(t)
			for _, name := range required {
				if !strings.Contains(diagnostic, name) {
					t.Errorf("missing diagnostic for %s", name)
				}
			}
			const secret = "test-only-private-value"
			t.Setenv("CPK_INTELLIGENCE_API_KEY", secret)
			if strings.Contains(startupError(t), secret) {
				t.Fatal("expected a secret-free configuration error")
			}
		})
	}
}

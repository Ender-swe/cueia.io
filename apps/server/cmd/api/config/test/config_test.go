package config_test

import (
	"net"
	"testing"

	"github.com/Ender-swe/cueia.io/apps/server/cmd/api/config"
)

func TestSampleDefaultListener(t *testing.T) {
	t.Setenv("CUEIA_MODE", "sample")
	t.Setenv("CUEIA_LISTEN_ADDR", "")
	t.Setenv("CPK_INTELLIGENCE_API_KEY", "test-only-private-value")
	settings, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	host, _, err := net.SplitHostPort(settings.Addr)
	if err != nil || !net.ParseIP(host).IsLoopback() {
		t.Fatal("default sample listener must bind to loopback")
	}
}

package config

import (
	"net"
	"testing"
)

func TestSampleDefaultListener(t *testing.T) {
	t.Setenv("CUEIA_MODE", "sample")
	t.Setenv("CUEIA_LISTEN_ADDR", "")
	t.Setenv("CPK_INTELLIGENCE_API_KEY", "test-only-private-value")
	settings, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	host, _, err := net.SplitHostPort(settings.Addr)
	if err != nil || !net.ParseIP(host).IsLoopback() {
		t.Fatal("default sample listener must bind to loopback")
	}
}

package main

import (
	"net"
	"os"
	"testing"
)

func TestSampleDefaultListener(t *testing.T) {
	t.Setenv("CUEIA_MODE", "sample")
	t.Setenv("CUEIA_LISTEN_ADDR", "")
	t.Setenv("CPK_INTELLIGENCE_API_KEY", "test-only-private-value")
	settings, err := loadAPIConfig()
	if err != nil {
		t.Fatal(err)
	}
	host, _, err := net.SplitHostPort(settings.Addr)
	if err != nil || !net.ParseIP(host).IsLoopback() {
		t.Fatal("default sample listener must bind to loopback")
	}
}

func browserClerkIssuer() string {
	return os.Getenv("CUEIA_BROWSER_CLERK_ISSUER")
}

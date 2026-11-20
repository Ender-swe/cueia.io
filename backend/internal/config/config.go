package config

import (
	"fmt"
	"os"
	"strings"
)

type API struct {
	Mode            string
	Addr            string
	ClerkIssuer     string
	AuthorizedParty string
}

func LoadAPI() (API, error) {
	c := API{Mode: os.Getenv("CUEIA_MODE"), Addr: os.Getenv("CUEIA_LISTEN_ADDR"), ClerkIssuer: os.Getenv("CLERK_ISSUER"), AuthorizedParty: os.Getenv("CLERK_AUTHORIZED_PARTY")}
	if c.Mode == "" {
		c.Mode = "sample"
	}
	if c.Mode != "sample" && c.Mode != "live" {
		return c, fmt.Errorf("invalid CUEIA_MODE")
	}
	var missing []string
	for _, k := range []string{"CLERK_ISSUER", "CLERK_AUTHORIZED_PARTY", "CPK_INTELLIGENCE_API_KEY"} {
		if strings.TrimSpace(os.Getenv(k)) == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return c, fmt.Errorf("missing settings: %s", strings.Join(missing, ", "))
	}
	if c.Addr == "" {
		c.Addr = "127.0.0.1:8080"
	}
	return c, nil
}

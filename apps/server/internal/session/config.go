package session

import (
	"os"
	"strings"
)

const (
	tokenBytes      = 32
	sessionLifetime = "12 hours"
)

type config struct {
	ClerkIssuer     string
	AuthorizedParty string
}

// loadConfig returns the session settings and the names of any missing variables.
func loadConfig() (config, []string) {
	c := config{
		ClerkIssuer:     os.Getenv("CLERK_ISSUER"),
		AuthorizedParty: os.Getenv("CLERK_AUTHORIZED_PARTY"),
	}
	var missing []string
	if strings.TrimSpace(c.ClerkIssuer) == "" {
		missing = append(missing, "CLERK_ISSUER")
	}
	if strings.TrimSpace(c.AuthorizedParty) == "" {
		missing = append(missing, "CLERK_AUTHORIZED_PARTY")
	}
	return c, missing
}

package session

import (
	"fmt"
	"os"
	"strings"
)

const (
	tokenBytes      = 32
	sessionLifetime = "12 hours"
)

type Config struct {
	ClerkIssuer     string
	AuthorizedParty string
}

// LoadConfig reads the required Clerk settings for the session service.
func LoadConfig() (Config, error) {
	c := Config{
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
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing settings: %s", strings.Join(missing, ", "))
	}
	return c, nil
}

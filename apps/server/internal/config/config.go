package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/Ender-swe/cueia.io/apps/server/internal/session"
)

type API struct {
	Mode    string
	Addr    string
	Session session.Config
}

func LoadAPI() (API, error) {
	sessions, missing := session.LoadConfig()
	c := API{Mode: os.Getenv("CUEIA_MODE"), Addr: os.Getenv("CUEIA_LISTEN_ADDR"), Session: sessions}
	if c.Mode == "" {
		c.Mode = "sample"
	}
	if c.Mode != "sample" && c.Mode != "live" {
		return c, fmt.Errorf("invalid CUEIA_MODE")
	}
	if strings.TrimSpace(os.Getenv("CPK_INTELLIGENCE_API_KEY")) == "" {
		missing = append(missing, "CPK_INTELLIGENCE_API_KEY")
	}
	if len(missing) > 0 {
		return c, fmt.Errorf("missing settings: %s", strings.Join(missing, ", "))
	}
	if c.Addr == "" {
		c.Addr = "127.0.0.1:8080"
	}
	return c, nil
}

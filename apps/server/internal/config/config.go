package config

import (
	"fmt"
	"os"
	"strings"
)

type API struct {
	Mode string
	Addr string
}

func LoadAPI() (API, error) {
	c := API{Mode: os.Getenv("CUEIA_MODE"), Addr: os.Getenv("CUEIA_LISTEN_ADDR")}
	if c.Mode == "" {
		c.Mode = "sample"
	}
	if c.Mode != "sample" && c.Mode != "live" {
		return c, fmt.Errorf("invalid CUEIA_MODE")
	}
	if strings.TrimSpace(os.Getenv("CPK_INTELLIGENCE_API_KEY")) == "" {
		return c, fmt.Errorf("missing settings: CPK_INTELLIGENCE_API_KEY")
	}
	if c.Addr == "" {
		c.Addr = "127.0.0.1:8080"
	}
	return c, nil
}

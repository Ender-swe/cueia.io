package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Ender-swe/cueia.io/apps/server/internal/config"
)

type apiConfig struct {
	Mode                   string
	Addr                   string
	DatabaseURL            string
	Deployment             string
	TestVolumeRoot         string
	AzureLocation          string
	TestPreviewClockOffset string
}

func loadAPIConfig() (apiConfig, error) {
	c := apiConfig{
		Mode:                   os.Getenv("CUEIA_MODE"),
		Addr:                   os.Getenv("CUEIA_LISTEN_ADDR"),
		DatabaseURL:            config.DatabaseURL(),
		Deployment:             config.DeploymentID(),
		TestVolumeRoot:         os.Getenv("CUEIA_TEST_VOLUME_ROOT"),
		AzureLocation:          config.AzureLocation(),
		TestPreviewClockOffset: os.Getenv("CUEIA_TEST_PREVIEW_CLOCK_OFFSET"),
	}
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

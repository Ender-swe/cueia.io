package config

import (
	"fmt"
	"os"
	"strings"

	sharedconfig "github.com/Ender-swe/cueia.io/apps/server/internal/config"
)

type Settings struct {
	Mode                   string
	Addr                   string
	DatabaseURL            string
	Deployment             string
	TestVolumeRoot         string
	AzureLocation          string
	TestPreviewClockOffset string
}

func Load() (Settings, error) {
	c := Settings{
		Mode:                   os.Getenv("CUEIA_MODE"),
		Addr:                   os.Getenv("CUEIA_LISTEN_ADDR"),
		DatabaseURL:            sharedconfig.DatabaseURL(),
		Deployment:             sharedconfig.DeploymentID(),
		TestVolumeRoot:         os.Getenv("CUEIA_TEST_VOLUME_ROOT"),
		AzureLocation:          sharedconfig.AzureLocation(),
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

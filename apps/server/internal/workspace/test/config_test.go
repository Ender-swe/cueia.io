package workspace_test

import "os"

type recoveryConfig struct {
	diskDir    string
	workspace  string
	deployment string
}

func loadRecoveryConfig() recoveryConfig {
	return recoveryConfig{
		diskDir:    os.Getenv("CUEIA_TEST_DISK_DIR"),
		workspace:  os.Getenv("CUEIA_TEST_REPLACE_WORKSPACE"),
		deployment: os.Getenv("COMPOSE_PROJECT_NAME"),
	}
}

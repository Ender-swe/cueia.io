package config

import "os"

func DatabaseURL() string {
	return os.Getenv("DATABASE_URL")
}

func TestDatabaseURL() string {
	return os.Getenv("CUEIA_TEST_DATABASE_URL")
}

func DeploymentID() string {
	return os.Getenv("CUEIA_DEPLOYMENT_ID")
}

func AzureLocation() string {
	return os.Getenv("AZURE_LOCATION")
}

func MigrationsDir(fallback string) string {
	if dir := os.Getenv("MIGRATIONS_DIR"); dir != "" {
		return dir
	}
	return fallback
}

package config

import (
	"log"
	"os"
)

// VMConfig holds the Virtual Machine configuration details
type VMConfig struct {
	Name            string
	Username        string
	Password        string
	SandboxUsername string
	SandBoxPassword string
}

// GetEnv fetches an environment variable or returns a default value
func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// LoadEnv validates that all required environment variables are set
func LoadEnv() {
	requiredVars := []string{
		"OPENAI_API_KEY", // For OpenAI integration
	}

	for _, envVar := range requiredVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("Environment variable %s is required but not set", envVar)
		}
	}
	log.Println("All required environment variables are successfully loaded.")
}

// Fetch specific environment variables with defaults
func GetOpenAIKey() string {
	return GetEnv("OPENAI_API_KEY", "")
}

func GetSensitiveDatasetPath() string {
	return GetEnv("BASELINE_DATASET", "./data/sensitive_data.db")
}

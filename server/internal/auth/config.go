package auth

import (
	"os"
)

// Config holds the Microsoft Graph / Entra ID configuration
type Config struct {
	ClientID     string
	ClientSecret string
	TenantID     string
	RedirectURL  string
	Scopes       []string
}

// NewConfigFromEnv creates a new Config from environment variables
func NewConfigFromEnv() *Config {
	// Default scopes for basic Graph API access
	scopes := []string{
		"User.Read",           // Read user profile
		"offline_access",      // Get refresh tokens
	}

	return &Config{
		ClientID:     getEnv("AZURE_CLIENT_ID", ""),
		ClientSecret: getEnv("AZURE_CLIENT_SECRET", ""),
		TenantID:     getEnv("AZURE_TENANT_ID", "common"),
		RedirectURL:  getEnv("AZURE_REDIRECT_URL", "http://localhost:8080/auth/callback"),
		Scopes:       scopes,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

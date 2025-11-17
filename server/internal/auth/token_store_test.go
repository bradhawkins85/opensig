package auth

import (
	"testing"
	"time"
)

func TestInMemoryTokenStore(t *testing.T) {
	store := NewInMemoryTokenStore()

	// Test storing a token
	tokenData := &TokenData{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    time.Now().Add(time.Hour),
		UserID:       "user123",
		Email:        "test@example.com",
	}

	err := store.Store("user123", tokenData)
	if err != nil {
		t.Fatalf("Failed to store token: %v", err)
	}

	// Test retrieving the token
	retrieved, err := store.Get("user123")
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	if retrieved.UserID != tokenData.UserID {
		t.Errorf("Expected UserID %s, got %s", tokenData.UserID, retrieved.UserID)
	}

	if retrieved.Email != tokenData.Email {
		t.Errorf("Expected Email %s, got %s", tokenData.Email, retrieved.Email)
	}

	if retrieved.AccessToken != tokenData.AccessToken {
		t.Errorf("Expected AccessToken %s, got %s", tokenData.AccessToken, retrieved.AccessToken)
	}

	// Test getting non-existent token
	_, err = store.Get("nonexistent")
	if err != ErrTokenNotFound {
		t.Errorf("Expected ErrTokenNotFound, got %v", err)
	}

	// Test deleting a token
	err = store.Delete("user123")
	if err != nil {
		t.Fatalf("Failed to delete token: %v", err)
	}

	// Verify token is gone
	_, err = store.Get("user123")
	if err != ErrTokenNotFound {
		t.Errorf("Expected ErrTokenNotFound after deletion, got %v", err)
	}
}

func TestConfig(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("AZURE_CLIENT_ID", "test-client-id")
	t.Setenv("AZURE_CLIENT_SECRET", "test-secret")
	t.Setenv("AZURE_TENANT_ID", "test-tenant")
	t.Setenv("AZURE_REDIRECT_URL", "http://localhost:8080/callback")

	config := NewConfigFromEnv()

	if config.ClientID != "test-client-id" {
		t.Errorf("Expected ClientID test-client-id, got %s", config.ClientID)
	}

	if config.ClientSecret != "test-secret" {
		t.Errorf("Expected ClientSecret test-secret, got %s", config.ClientSecret)
	}

	if config.TenantID != "test-tenant" {
		t.Errorf("Expected TenantID test-tenant, got %s", config.TenantID)
	}

	if config.RedirectURL != "http://localhost:8080/callback" {
		t.Errorf("Expected RedirectURL http://localhost:8080/callback, got %s", config.RedirectURL)
	}

	// Verify default scopes
	expectedScopes := []string{"User.Read", "offline_access"}
	if len(config.Scopes) != len(expectedScopes) {
		t.Errorf("Expected %d scopes, got %d", len(expectedScopes), len(config.Scopes))
	}

	for i, scope := range expectedScopes {
		if config.Scopes[i] != scope {
			t.Errorf("Expected scope %s at index %d, got %s", scope, i, config.Scopes[i])
		}
	}
}

func TestConfigDefaults(t *testing.T) {
	// Don't set any environment variables, test defaults
	config := &Config{
		ClientID:     "",
		ClientSecret: "",
		TenantID:     "",
		RedirectURL:  "",
		Scopes:       []string{},
	}

	// These should be populated with defaults by NewConfigFromEnv
	config = NewConfigFromEnv()

	if config.TenantID != "common" {
		t.Errorf("Expected default TenantID common, got %s", config.TenantID)
	}

	if config.RedirectURL != "http://localhost:8080/auth/callback" {
		t.Errorf("Expected default RedirectURL, got %s", config.RedirectURL)
	}

	// Client ID can be empty but should be set from env or empty string
	if config.ClientID == "" {
		t.Log("ClientID is empty (expected when AZURE_CLIENT_ID not set)")
	}
}

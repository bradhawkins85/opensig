package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
	"github.com/google/uuid"
)

// Service handles Microsoft Graph authentication
type Service struct {
	config                 *Config
	tokenStore             TokenStore
	hasConfidentialClient  bool
	// For web flow
	confidentialClient confidential.Client
	// For device code flow
	publicClient public.Client
	// State management for OAuth flow
	stateStore map[string]bool
}

// NewService creates a new auth service
func NewService(config *Config, tokenStore TokenStore) (*Service, error) {
	if config.ClientID == "" {
		return nil, errors.New("AZURE_CLIENT_ID is required")
	}

	authority := fmt.Sprintf("https://login.microsoftonline.com/%s", config.TenantID)

	var confidentialClient confidential.Client
	var publicClient public.Client
	var err error
	hasConfidentialClient := false

	// Initialize confidential client for web flow (if client secret is provided)
	if config.ClientSecret != "" {
		cred, err := confidential.NewCredFromSecret(config.ClientSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to create credential: %w", err)
		}

		confidentialClient, err = confidential.New(authority, config.ClientID, cred,
			confidential.WithHTTPClient(&http.Client{}))
		if err != nil {
			return nil, fmt.Errorf("failed to create confidential client: %w", err)
		}
		hasConfidentialClient = true
	}

	// Initialize public client for device code flow
	publicClient, err = public.New(config.ClientID, public.WithAuthority(authority))
	if err != nil {
		return nil, fmt.Errorf("failed to create public client: %w", err)
	}

	return &Service{
		config:                config,
		tokenStore:            tokenStore,
		hasConfidentialClient: hasConfidentialClient,
		confidentialClient:    confidentialClient,
		publicClient:          publicClient,
		stateStore:            make(map[string]bool),
	}, nil
}

// GetAuthURL generates the Microsoft login URL for OAuth flow
func (s *Service) GetAuthURL() (string, string, error) {
	if !s.hasConfidentialClient {
		return "", "", errors.New("confidential client not initialized - AZURE_CLIENT_SECRET required")
	}

	// Generate a random state parameter to prevent CSRF
	state := uuid.New().String()
	s.stateStore[state] = true

	// Build the authorization URL
	authURL, err := s.confidentialClient.AuthCodeURL(context.Background(), s.config.ClientID, 
		s.config.RedirectURL, s.config.Scopes)
	if err != nil {
		return "", "", fmt.Errorf("failed to create auth URL: %w", err)
	}

	// Append state parameter
	authURL = authURL + "&state=" + state

	return authURL, state, nil
}

// ValidateState checks if the state parameter is valid
func (s *Service) ValidateState(state string) bool {
	if _, ok := s.stateStore[state]; ok {
		delete(s.stateStore, state) // One-time use
		return true
	}
	return false
}

// ExchangeCodeForToken exchanges an authorization code for tokens
func (s *Service) ExchangeCodeForToken(ctx context.Context, code string) (*TokenData, error) {
	if !s.hasConfidentialClient {
		return nil, errors.New("confidential client not initialized")
	}

	result, err := s.confidentialClient.AcquireTokenByAuthCode(ctx, code, s.config.RedirectURL, s.config.Scopes)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire token: %w", err)
	}

	// Get user info from the access token
	userInfo, err := s.getUserInfo(result.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	tokenData := &TokenData{
		AccessToken:  result.AccessToken,
		RefreshToken: "", // MSAL handles refresh internally
		ExpiresAt:    result.ExpiresOn,
		UserID:       userInfo.ID,
		Email:        userInfo.Email,
	}

	// Store the token
	if err := s.tokenStore.Store(userInfo.ID, tokenData); err != nil {
		return nil, fmt.Errorf("failed to store token: %w", err)
	}

	return tokenData, nil
}

// DeviceCodeFlow initiates device code flow for local development
func (s *Service) DeviceCodeFlow(ctx context.Context) (*TokenData, error) {
	dc, err := s.publicClient.AcquireTokenByDeviceCode(ctx, s.config.Scopes)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate device code flow: %w", err)
	}

	// Display the device code message to the user
	fmt.Println(dc.Result.Message)

	result, err := dc.AuthenticationResult(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire token via device code: %w", err)
	}

	// Get user info
	userInfo, err := s.getUserInfo(result.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	tokenData := &TokenData{
		AccessToken:  result.AccessToken,
		RefreshToken: "", // MSAL handles refresh internally
		ExpiresAt:    result.ExpiresOn,
		UserID:       userInfo.ID,
		Email:        userInfo.Email,
	}

	// Store the token
	if err := s.tokenStore.Store(userInfo.ID, tokenData); err != nil {
		return nil, fmt.Errorf("failed to store token: %w", err)
	}

	return tokenData, nil
}

// UserInfo represents basic user information from Microsoft Graph
type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"userPrincipalName"`
}

// getUserInfo fetches user information from Microsoft Graph
func (s *Service) getUserInfo(accessToken string) (*UserInfo, error) {
	req, err := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me?$select=id,userPrincipalName", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("graph API returned %d: %s", resp.StatusCode, string(body))
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// GetToken retrieves a stored token for a user
func (s *Service) GetToken(userID string) (*TokenData, error) {
	return s.tokenStore.Get(userID)
}

// DeleteToken removes a stored token for a user
func (s *Service) DeleteToken(userID string) error {
	return s.tokenStore.Delete(userID)
}

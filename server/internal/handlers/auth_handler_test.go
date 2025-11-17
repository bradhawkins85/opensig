package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/opensig/server/internal/auth"
)

// mockAuthService is a mock implementation of AuthService for testing
type mockAuthService struct {
	authURL       string
	state         string
	authURLError  error
	validateState bool
	tokenData     *auth.TokenData
	tokenError    error
}

func (m *mockAuthService) GetAuthURL() (string, string, error) {
	return m.authURL, m.state, m.authURLError
}

func (m *mockAuthService) ValidateState(state string) bool {
	return m.validateState
}

func (m *mockAuthService) ExchangeCodeForToken(ctx context.Context, code string) (*auth.TokenData, error) {
	return m.tokenData, m.tokenError
}

func (m *mockAuthService) GetToken(userID string) (*auth.TokenData, error) {
	return m.tokenData, m.tokenError
}

func (m *mockAuthService) DeleteToken(userID string) error {
	return m.tokenError
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockService := &mockAuthService{
		authURL: "https://login.microsoftonline.com/authorize?client_id=test",
		state:   "test-state",
	}

	handler := &AuthHandler{authService: mockService}

	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != mockService.authURL {
		t.Errorf("Expected redirect to %s, got %s", mockService.authURL, location)
	}
}

func TestAuthHandler_Callback_Success(t *testing.T) {
	tokenData := &auth.TokenData{
		UserID:       "user123",
		Email:        "test@example.com",
		AccessToken:  "token123",
		RefreshToken: "refresh123",
		ExpiresAt:    time.Now().Add(time.Hour),
	}

	mockService := &mockAuthService{
		validateState: true,
		tokenData:     tokenData,
	}

	handler := &AuthHandler{authService: mockService}

	req := httptest.NewRequest(http.MethodGet, "/auth/callback?code=authcode123&state=valid-state", nil)
	w := httptest.NewRecorder()

	handler.Callback(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["message"] != "Authentication successful" {
		t.Errorf("Expected success message, got %v", response["message"])
	}
}

func TestAuthHandler_Callback_MissingCode(t *testing.T) {
	mockService := &mockAuthService{}
	handler := &AuthHandler{authService: mockService}

	req := httptest.NewRequest(http.MethodGet, "/auth/callback?state=valid-state", nil)
	w := httptest.NewRecorder()

	handler.Callback(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthHandler_Callback_InvalidState(t *testing.T) {
	mockService := &mockAuthService{
		validateState: false,
	}
	handler := &AuthHandler{authService: mockService}

	req := httptest.NewRequest(http.MethodGet, "/auth/callback?code=authcode123&state=invalid-state", nil)
	w := httptest.NewRecorder()

	handler.Callback(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestAuthHandler_Status_Authenticated(t *testing.T) {
	tokenData := &auth.TokenData{
		UserID:      "user123",
		Email:       "test@example.com",
		AccessToken: "token123",
		ExpiresAt:   time.Now().Add(time.Hour),
	}

	mockService := &mockAuthService{
		tokenData: tokenData,
	}

	handler := &AuthHandler{authService: mockService}

	req := httptest.NewRequest(http.MethodGet, "/auth/status?user_id=user123", nil)
	w := httptest.NewRecorder()

	handler.Status(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["authenticated"] != true {
		t.Errorf("Expected authenticated to be true")
	}
}

func TestAuthHandler_Status_NotAuthenticated(t *testing.T) {
	mockService := &mockAuthService{
		tokenError: auth.ErrTokenNotFound,
	}

	handler := &AuthHandler{authService: mockService}

	req := httptest.NewRequest(http.MethodGet, "/auth/status?user_id=user123", nil)
	w := httptest.NewRecorder()

	handler.Status(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["authenticated"] != false {
		t.Errorf("Expected authenticated to be false")
	}
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	mockService := &mockAuthService{}
	handler := &AuthHandler{authService: mockService}

	body := bytes.NewBufferString(`{"user_id":"user123"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", body)
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["message"] != "Logged out successfully" {
		t.Errorf("Expected success message, got %s", response["message"])
	}
}

package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/your-org/opensig/server/internal/auth"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	GetAuthURL() (string, string, error)
	ValidateState(state string) bool
	ExchangeCodeForToken(ctx context.Context, code string) (*auth.TokenData, error)
	GetToken(userID string) (*auth.TokenData, error)
	DeleteToken(userID string) error
}

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login initiates the OAuth login flow by redirecting to Microsoft
// GET /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authURL, state, err := h.authService.GetAuthURL()
	if err != nil {
		log.Printf("Error generating auth URL: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to initiate login",
		})
		return
	}

	log.Printf("Redirecting to Microsoft login (state: %s)", state)

	// Redirect to Microsoft sign-in
	http.Redirect(w, r, authURL, http.StatusFound)
}

// Callback handles the OAuth callback from Microsoft
// GET /auth/callback?code=...&state=...
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get code and state from query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Missing authorization code",
		})
		return
	}

	// Validate state to prevent CSRF
	if !h.authService.ValidateState(state) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid state parameter",
		})
		return
	}

	// Exchange code for token
	tokenData, err := h.authService.ExchangeCodeForToken(r.Context(), code)
	if err != nil {
		log.Printf("Error exchanging code for token: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to complete authentication",
		})
		return
	}

	log.Printf("User authenticated successfully: %s (%s)", tokenData.Email, tokenData.UserID)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Authentication successful",
		"user": map[string]string{
			"id":    tokenData.UserID,
			"email": tokenData.Email,
		},
	})
}

// Status checks the authentication status for a user
// GET /auth/status?user_id=...
func (h *AuthHandler) Status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "user_id parameter is required",
		})
		return
	}

	tokenData, err := h.authService.GetToken(userID)
	if err != nil {
		if err == auth.ErrTokenNotFound {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"authenticated": false,
				"message":       "No active session",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to check status",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"authenticated": true,
		"user": map[string]string{
			"id":    tokenData.UserID,
			"email": tokenData.Email,
		},
		"expires_at": tokenData.ExpiresAt,
	})
}

// Logout removes the authentication token for a user
// POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	if req.UserID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "user_id is required",
		})
		return
	}

	if err := h.authService.DeleteToken(req.UserID); err != nil {
		log.Printf("Error deleting token: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to logout",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

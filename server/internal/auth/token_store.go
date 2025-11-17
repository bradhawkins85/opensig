package auth

import (
	"errors"
	"sync"
	"time"
)

var (
	// ErrTokenNotFound is returned when a token is not found in the store
	ErrTokenNotFound = errors.New("token not found")
)

// TokenData represents stored authentication tokens
type TokenData struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	UserID       string    `json:"user_id"`
	Email        string    `json:"email"`
}

// TokenStore defines the interface for storing and retrieving tokens
type TokenStore interface {
	Store(userID string, token *TokenData) error
	Get(userID string) (*TokenData, error)
	Delete(userID string) error
}

// InMemoryTokenStore is a simple in-memory token store for development
// WARNING: This is NOT suitable for production use - tokens will be lost on restart
type InMemoryTokenStore struct {
	mu     sync.RWMutex
	tokens map[string]*TokenData
}

// NewInMemoryTokenStore creates a new in-memory token store
func NewInMemoryTokenStore() *InMemoryTokenStore {
	return &InMemoryTokenStore{
		tokens: make(map[string]*TokenData),
	}
}

// Store saves a token for a user
func (s *InMemoryTokenStore) Store(userID string, token *TokenData) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[userID] = token
	return nil
}

// Get retrieves a token for a user
func (s *InMemoryTokenStore) Get(userID string) (*TokenData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	token, ok := s.tokens[userID]
	if !ok {
		return nil, ErrTokenNotFound
	}
	
	return token, nil
}

// Delete removes a token for a user
func (s *InMemoryTokenStore) Delete(userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, userID)
	return nil
}

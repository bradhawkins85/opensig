package store

import (
	"errors"
	"sync"
	"time"

	"github.com/your-org/opensig/server/internal/models"
)

var (
	ErrTenantNotFound     = errors.New("tenant not found")
	ErrTenantAlreadyExists = errors.New("tenant already exists")
)

// TenantStore provides in-memory storage for tenants
type TenantStore struct {
	mu      sync.RWMutex
	tenants map[string]*models.Tenant
}

// NewTenantStore creates a new tenant store
func NewTenantStore() *TenantStore {
	return &TenantStore{
		tenants: make(map[string]*models.Tenant),
	}
}

// Create adds a new tenant
func (s *TenantStore) Create(tenant *models.Tenant) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tenants[tenant.ID]; exists {
		return ErrTenantAlreadyExists
	}

	now := time.Now().UTC()
	tenant.CreatedAt = now
	tenant.UpdatedAt = now
	s.tenants[tenant.ID] = tenant
	return nil
}

// Get retrieves a tenant by ID
func (s *TenantStore) Get(id string) (*models.Tenant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tenant, exists := s.tenants[id]
	if !exists {
		return nil, ErrTenantNotFound
	}
	return tenant, nil
}

// List returns all tenants
func (s *TenantStore) List() []*models.Tenant {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.Tenant, 0, len(s.tenants))
	for _, tenant := range s.tenants {
		result = append(result, tenant)
	}
	return result
}

// Update modifies an existing tenant
func (s *TenantStore) Update(tenant *models.Tenant) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tenants[tenant.ID]; !exists {
		return ErrTenantNotFound
	}

	tenant.UpdatedAt = time.Now().UTC()
	s.tenants[tenant.ID] = tenant
	return nil
}

// Delete removes a tenant by ID
func (s *TenantStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tenants[id]; !exists {
		return ErrTenantNotFound
	}

	delete(s.tenants, id)
	return nil
}

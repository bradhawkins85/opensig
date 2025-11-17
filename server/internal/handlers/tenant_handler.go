package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/your-org/opensig/server/internal/models"
	"github.com/your-org/opensig/server/internal/store"
)

// TenantHandler handles tenant CRUD operations
type TenantHandler struct {
	store *store.TenantStore
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(store *store.TenantStore) *TenantHandler {
	return &TenantHandler{store: store}
}

// CreateTenant handles POST /v1/tenants
func (h *TenantHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tenant models.Tenant
	if err := json.NewDecoder(r.Body).Decode(&tenant); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	if err := h.store.Create(&tenant); err != nil {
		w.Header().Set("Content-Type", "application/json")
		if err == store.ErrTenantAlreadyExists {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to create tenant"})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(tenant)
}

// GetTenant handles GET /v1/tenants/{id}
func (h *TenantHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path (simple parsing since we're not using a router)
	id := r.URL.Path[len("/v1/tenants/"):]
	if id == "" {
		http.Error(w, "tenant id required", http.StatusBadRequest)
		return
	}

	tenant, err := h.store.Get(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if err == store.ErrTenantNotFound {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to get tenant"})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tenant)
}

// ListTenants handles GET /v1/tenants
func (h *TenantHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tenants := h.store.List()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"tenants": tenants,
		"count":   len(tenants),
	})
}

// UpdateTenant handles PUT /v1/tenants/{id}
func (h *TenantHandler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[len("/v1/tenants/"):]
	if id == "" {
		http.Error(w, "tenant id required", http.StatusBadRequest)
		return
	}

	var tenant models.Tenant
	if err := json.NewDecoder(r.Body).Decode(&tenant); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	tenant.ID = id
	if err := h.store.Update(&tenant); err != nil {
		w.Header().Set("Content-Type", "application/json")
		if err == store.ErrTenantNotFound {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to update tenant"})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tenant)
}

// DeleteTenant handles DELETE /v1/tenants/{id}
func (h *TenantHandler) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[len("/v1/tenants/"):]
	if id == "" {
		http.Error(w, "tenant id required", http.StatusBadRequest)
		return
	}

	if err := h.store.Delete(id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		if err == store.ErrTenantNotFound {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "failed to delete tenant"})
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

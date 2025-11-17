package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/opensig/server/internal/models"
	"github.com/your-org/opensig/server/internal/store"
)

func TestCreateTenant(t *testing.T) {
	store := store.NewTenantStore()
	handler := NewTenantHandler(store)

	tenant := models.Tenant{
		ID:     "tenant-1",
		Name:   "Test Tenant",
		Domain: "test.com",
		Active: true,
	}

	body, _ := json.Marshal(tenant)
	req := httptest.NewRequest(http.MethodPost, "/v1/tenants", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.CreateTenant(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var response models.Tenant
	_ = json.NewDecoder(rec.Body).Decode(&response)
	if response.ID != tenant.ID {
		t.Errorf("expected tenant ID %s, got %s", tenant.ID, response.ID)
	}
}

func TestCreateTenant_InvalidBody(t *testing.T) {
	store := store.NewTenantStore()
	handler := NewTenantHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/v1/tenants", bytes.NewReader([]byte("invalid json")))
	rec := httptest.NewRecorder()

	handler.CreateTenant(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCreateTenant_Duplicate(t *testing.T) {
	store := store.NewTenantStore()
	handler := NewTenantHandler(store)

	tenant := models.Tenant{ID: "tenant-1", Name: "Test", Domain: "test.com", Active: true}
	_ = store.Create(&tenant)

	body, _ := json.Marshal(tenant)
	req := httptest.NewRequest(http.MethodPost, "/v1/tenants", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.CreateTenant(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
}

func TestGetTenant(t *testing.T) {
	store := store.NewTenantStore()
	handler := NewTenantHandler(store)

	tenant := models.Tenant{ID: "tenant-1", Name: "Test", Domain: "test.com", Active: true}
	_ = store.Create(&tenant)

	req := httptest.NewRequest(http.MethodGet, "/v1/tenants/tenant-1", nil)
	rec := httptest.NewRecorder()

	handler.GetTenant(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response models.Tenant
	_ = json.NewDecoder(rec.Body).Decode(&response)
	if response.ID != tenant.ID {
		t.Errorf("expected tenant ID %s, got %s", tenant.ID, response.ID)
	}
}

func TestGetTenant_NotFound(t *testing.T) {
	store := store.NewTenantStore()
	handler := NewTenantHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/v1/tenants/nonexistent", nil)
	rec := httptest.NewRecorder()

	handler.GetTenant(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestListTenants(t *testing.T) {
	store := store.NewTenantStore()
	handler := NewTenantHandler(store)

	tenant1 := models.Tenant{ID: "tenant-1", Name: "Test 1", Domain: "t1.com", Active: true}
	tenant2 := models.Tenant{ID: "tenant-2", Name: "Test 2", Domain: "t2.com", Active: true}
	_ = store.Create(&tenant1)
	_ = store.Create(&tenant2)

	req := httptest.NewRequest(http.MethodGet, "/v1/tenants", nil)
	rec := httptest.NewRecorder()

	handler.ListTenants(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&response)
	count := int(response["count"].(float64))
	if count != 2 {
		t.Errorf("expected 2 tenants, got %d", count)
	}
}

func TestUpdateTenant(t *testing.T) {
	store := store.NewTenantStore()
	handler := NewTenantHandler(store)

	tenant := models.Tenant{ID: "tenant-1", Name: "Test", Domain: "test.com", Active: true}
	_ = store.Create(&tenant)

	tenant.Name = "Updated Test"
	body, _ := json.Marshal(tenant)
	req := httptest.NewRequest(http.MethodPut, "/v1/tenants/tenant-1", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.UpdateTenant(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response models.Tenant
	_ = json.NewDecoder(rec.Body).Decode(&response)
	if response.Name != "Updated Test" {
		t.Errorf("expected name 'Updated Test', got %s", response.Name)
	}
}

func TestDeleteTenant(t *testing.T) {
	store := store.NewTenantStore()
	handler := NewTenantHandler(store)

	tenant := models.Tenant{ID: "tenant-1", Name: "Test", Domain: "test.com", Active: true}
	_ = store.Create(&tenant)

	req := httptest.NewRequest(http.MethodDelete, "/v1/tenants/tenant-1", nil)
	rec := httptest.NewRecorder()

	handler.DeleteTenant(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	// Verify deletion
	_, err := store.Get("tenant-1")
	if err == nil {
		t.Error("expected tenant to be deleted")
	}
}

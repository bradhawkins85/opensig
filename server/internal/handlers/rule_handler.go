package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/your-org/opensig/server/internal/models"
	"github.com/your-org/opensig/server/internal/store"
)

// RuleHandler handles rule-related HTTP requests
type RuleHandler struct {
	ruleStore *store.RuleStore
}

// NewRuleHandler creates a new rule handler
func NewRuleHandler(ruleStore *store.RuleStore) *RuleHandler {
	return &RuleHandler{
		ruleStore: ruleStore,
	}
}

// CreateRule handles POST /v1/rules
func (h *RuleHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	var rule models.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.ruleStore.CreateRule(&rule); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

// GetRule handles GET /v1/rules/{id}
func (h *RuleHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/rules/")
	if id == "" {
		http.Error(w, "Rule ID required", http.StatusBadRequest)
		return
	}

	rule, err := h.ruleStore.GetRule(id)
	if err != nil {
		if err == store.ErrRuleNotFound {
			http.Error(w, "Rule not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}

// ListRules handles GET /v1/rules?tenant_id={tenantID}
func (h *RuleHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		http.Error(w, "tenant_id query parameter required", http.StatusBadRequest)
		return
	}

	rules, err := h.ruleStore.GetRulesByTenantID(tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rules)
}

// UpdateRule handles PUT /v1/rules/{id}
func (h *RuleHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/rules/")
	if id == "" {
		http.Error(w, "Rule ID required", http.StatusBadRequest)
		return
	}

	var rule models.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	rule.ID = id
	if err := h.ruleStore.UpdateRule(&rule); err != nil {
		if err == store.ErrRuleNotFound {
			http.Error(w, "Rule not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rule)
}

// DeleteRule handles DELETE /v1/rules/{id}
func (h *RuleHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/rules/")
	if id == "" {
		http.Error(w, "Rule ID required", http.StatusBadRequest)
		return
	}

	if err := h.ruleStore.DeleteRule(id); err != nil {
		if err == store.ErrRuleNotFound {
			http.Error(w, "Rule not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// EvaluateRules handles POST /v1/rules/evaluate
func (h *RuleHandler) EvaluateRules(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantID string              `json:"tenant_id"`
		Message  models.TestMessage  `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TenantID == "" {
		http.Error(w, "tenant_id is required", http.StatusBadRequest)
		return
	}

	result, err := h.ruleStore.EvaluateRules(req.TenantID, &req.Message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

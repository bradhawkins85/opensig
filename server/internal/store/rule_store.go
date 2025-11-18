package store

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/opensig/server/internal/models"
)

var (
	ErrRuleNotFound = errors.New("rule not found")
)

// RuleStore manages rule data
type RuleStore struct {
	mu    sync.RWMutex
	rules map[string]*models.Rule
	scheduleStore *ScheduleStore // Reference to schedule store for schedule evaluation
}

// NewRuleStore creates a new rule store
func NewRuleStore(scheduleStore *ScheduleStore) *RuleStore {
	return &RuleStore{
		rules: make(map[string]*models.Rule),
		scheduleStore: scheduleStore,
	}
}

// CreateRule creates a new rule
func (s *RuleStore) CreateRule(rule *models.Rule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}
	rule.CreatedAt = time.Now().UTC()
	rule.UpdatedAt = time.Now().UTC()

	s.rules[rule.ID] = rule
	return nil
}

// GetRule retrieves a rule by ID
func (s *RuleStore) GetRule(id string) (*models.Rule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rule, exists := s.rules[id]
	if !exists {
		return nil, ErrRuleNotFound
	}

	return rule, nil
}

// GetRulesByTenantID returns all rules for a tenant
func (s *RuleStore) GetRulesByTenantID(tenantID string) ([]*models.Rule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*models.Rule
	for _, rule := range s.rules {
		if rule.TenantID == tenantID {
			result = append(result, rule)
		}
	}

	// Sort by priority (descending)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Priority > result[j].Priority
	})

	return result, nil
}

// UpdateRule updates an existing rule
func (s *RuleStore) UpdateRule(rule *models.Rule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.rules[rule.ID]; !exists {
		return ErrRuleNotFound
	}

	rule.UpdatedAt = time.Now().UTC()
	s.rules[rule.ID] = rule
	return nil
}

// DeleteRule deletes a rule by ID
func (s *RuleStore) DeleteRule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.rules[id]; !exists {
		return ErrRuleNotFound
	}

	delete(s.rules, id)
	return nil
}

// EvaluateRules evaluates all active rules for a tenant against a test message
func (s *RuleStore) EvaluateRules(tenantID string, message *models.TestMessage) (*models.RuleEvaluationResult, error) {
	rules, err := s.GetRulesByTenantID(tenantID)
	if err != nil {
		return nil, err
	}

	result := &models.RuleEvaluationResult{
		MatchedRules: []models.Rule{},
	}

	// Evaluate each rule
	for _, rule := range rules {
		if !rule.Active {
			continue
		}

		// Check if schedule allows (if schedule is set)
		if rule.ScheduleID != "" && s.scheduleStore != nil {
			schedule, err := s.scheduleStore.GetSchedule(rule.ScheduleID)
			if err == nil && schedule.Active {
				if !s.scheduleStore.IsTimeInSchedule(schedule, message.Timestamp) {
					continue // Skip this rule if time is not in schedule
				}
			}
		}

		// Evaluate conditions
		if s.evaluateConditions(&rule.Conditions, message) {
			result.MatchedRules = append(result.MatchedRules, *rule)

			// If this is an exclusive rule, it's the selected rule and we stop
			if rule.Exclusive {
				result.SelectedRule = rule
				break
			}
		}
	}

	// If no exclusive rule matched, select the highest priority matched rule
	if result.SelectedRule == nil && len(result.MatchedRules) > 0 {
		result.SelectedRule = &result.MatchedRules[0]
	}

	return result, nil
}

// evaluateConditions checks if a message matches the rule conditions
func (s *RuleStore) evaluateConditions(conditions *models.Conditions, message *models.TestMessage) bool {
	// If no conditions are specified, rule matches everything
	hasConditions := len(conditions.SenderEmails) > 0 ||
		len(conditions.SenderDomains) > 0 ||
		len(conditions.RecipientEmails) > 0 ||
		len(conditions.RecipientDomains) > 0 ||
		len(conditions.MessageTypes) > 0

	if !hasConditions {
		return true
	}

	// Check sender email
	if len(conditions.SenderEmails) > 0 {
		matched := false
		for _, email := range conditions.SenderEmails {
			if strings.EqualFold(email, message.SenderEmail) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check sender domain
	if len(conditions.SenderDomains) > 0 {
		senderDomain := extractDomain(message.SenderEmail)
		matched := false
		for _, domain := range conditions.SenderDomains {
			if strings.EqualFold(domain, senderDomain) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check recipient emails
	if len(conditions.RecipientEmails) > 0 {
		matched := false
		for _, condEmail := range conditions.RecipientEmails {
			for _, recipientEmail := range message.RecipientEmails {
				if strings.EqualFold(condEmail, recipientEmail) {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check recipient domains
	if len(conditions.RecipientDomains) > 0 {
		matched := false
		for _, condDomain := range conditions.RecipientDomains {
			for _, recipientEmail := range message.RecipientEmails {
				recipientDomain := extractDomain(recipientEmail)
				if strings.EqualFold(condDomain, recipientDomain) {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check message type
	if len(conditions.MessageTypes) > 0 {
		matched := false
		for _, msgType := range conditions.MessageTypes {
			if msgType == message.MessageType {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

// extractDomain extracts the domain from an email address
func extractDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

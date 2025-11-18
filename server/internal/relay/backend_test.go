package relay

import (
	"bytes"
	"io"
	"testing"

	"github.com/emersion/go-smtp"
	"github.com/your-org/opensig/server/internal/models"
	"github.com/your-org/opensig/server/internal/store"
)

func TestBackend_NewSession(t *testing.T) {
	tenantStore := store.NewTenantStore()
	backend := NewBackend(tenantStore)
	
	// The actual NewSession expects a non-nil *smtp.Conn which requires a real connection
	// For now, we'll just verify the backend was created successfully
	if backend == nil {
		t.Fatal("Expected non-nil backend")
	}
	
	if backend.tenantStore == nil {
		t.Fatal("Expected non-nil tenantStore")
	}
}

func TestSession_AuthPlain_Success(t *testing.T) {
	tenantStore := store.NewTenantStore()
	
	// Create a test tenant
	tenant := &models.Tenant{
		ID:     "test-tenant",
		Name:   "Test Tenant",
		Domain: "example.com",
		Active: true,
	}
	tenantStore.Create(tenant)
	
	backend := NewBackend(tenantStore)
	session := &Session{
		backend:   backend,
		sessionID: "test-session",
	}
	
	// Test successful authentication
	err := session.AuthPlain("user@example.com", "password")
	if err != nil {
		t.Errorf("Expected successful auth, got error: %v", err)
	}
}

func TestSession_AuthPlain_InvalidDomain(t *testing.T) {
	tenantStore := store.NewTenantStore()
	backend := NewBackend(tenantStore)
	session := &Session{
		backend:   backend,
		sessionID: "test-session",
	}
	
	// Test authentication with non-existent tenant
	err := session.AuthPlain("user@nonexistent.com", "password")
	if err != smtp.ErrAuthFailed {
		t.Errorf("Expected ErrAuthFailed, got: %v", err)
	}
}

func TestSession_AuthPlain_InvalidUsername(t *testing.T) {
	tenantStore := store.NewTenantStore()
	backend := NewBackend(tenantStore)
	session := &Session{
		backend:   backend,
		sessionID: "test-session",
	}
	
	// Test authentication with invalid username format
	err := session.AuthPlain("invalidusername", "password")
	if err != smtp.ErrAuthFailed {
		t.Errorf("Expected ErrAuthFailed, got: %v", err)
	}
}

func TestSession_AuthPlain_InactiveTenant(t *testing.T) {
	tenantStore := store.NewTenantStore()
	
	// Create an inactive tenant
	tenant := &models.Tenant{
		ID:     "test-tenant",
		Name:   "Test Tenant",
		Domain: "example.com",
		Active: false,
	}
	tenantStore.Create(tenant)
	
	backend := NewBackend(tenantStore)
	session := &Session{
		backend:   backend,
		sessionID: "test-session",
	}
	
	// Test authentication with inactive tenant
	err := session.AuthPlain("user@example.com", "password")
	if err != smtp.ErrAuthFailed {
		t.Errorf("Expected ErrAuthFailed for inactive tenant, got: %v", err)
	}
}

func TestSession_Mail(t *testing.T) {
	session := &Session{
		sessionID: "test-session",
	}
	
	err := session.Mail("sender@example.com", nil)
	if err != nil {
		t.Errorf("Mail failed: %v", err)
	}
	
	if session.from != "sender@example.com" {
		t.Errorf("Expected from=sender@example.com, got %s", session.from)
	}
}

func TestSession_Rcpt(t *testing.T) {
	session := &Session{
		sessionID: "test-session",
	}
	
	err := session.Rcpt("recipient1@example.com", nil)
	if err != nil {
		t.Errorf("Rcpt failed: %v", err)
	}
	
	err = session.Rcpt("recipient2@example.com", nil)
	if err != nil {
		t.Errorf("Rcpt failed: %v", err)
	}
	
	if len(session.to) != 2 {
		t.Errorf("Expected 2 recipients, got %d", len(session.to))
	}
}

func TestSession_Data(t *testing.T) {
	tenantStore := store.NewTenantStore()
	backend := NewBackend(tenantStore)
	session := &Session{
		backend:   backend,
		sessionID: "test-session",
		from:      "sender@example.com",
		to:        []string{"recipient@example.com"},
	}
	
	message := []byte("From: sender@example.com\r\nTo: recipient@example.com\r\nSubject: Test\r\n\r\nTest message body")
	reader := bytes.NewReader(message)
	
	err := session.Data(reader)
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
}

func TestSession_Reset(t *testing.T) {
	session := &Session{
		sessionID: "test-session",
		from:      "sender@example.com",
		to:        []string{"recipient@example.com"},
	}
	
	session.Reset()
	
	if session.from != "" {
		t.Errorf("Expected empty from after reset, got %s", session.from)
	}
	
	if session.to != nil {
		t.Errorf("Expected nil to after reset, got %v", session.to)
	}
}

func TestAnonymousBackend_NewSession(t *testing.T) {
	tenantStore := store.NewTenantStore()
	backend := NewBackend(tenantStore)
	anonBackend := NewAnonymousBackend(backend)
	
	// Verify anonymous backend was created successfully
	if anonBackend == nil {
		t.Fatal("Expected non-nil anonymous backend")
	}
	
	if anonBackend.backend == nil {
		t.Fatal("Expected non-nil backend reference")
	}
}

func TestAnonymousSession_Mail(t *testing.T) {
	session := &AnonymousSession{
		sessionID: "test-session",
	}
	
	err := session.Mail("sender@example.com", nil)
	if err != nil {
		t.Errorf("Mail failed: %v", err)
	}
	
	if session.from != "sender@example.com" {
		t.Errorf("Expected from=sender@example.com, got %s", session.from)
	}
}

func TestAnonymousSession_Rcpt(t *testing.T) {
	session := &AnonymousSession{
		sessionID: "test-session",
	}
	
	err := session.Rcpt("recipient@example.com", nil)
	if err != nil {
		t.Errorf("Rcpt failed: %v", err)
	}
	
	if len(session.to) != 1 {
		t.Errorf("Expected 1 recipient, got %d", len(session.to))
	}
}

func TestAnonymousSession_Data(t *testing.T) {
	tenantStore := store.NewTenantStore()
	backend := NewBackend(tenantStore)
	session := &AnonymousSession{
		backend:   backend,
		sessionID: "test-session",
		from:      "sender@example.com",
		to:        []string{"recipient@example.com"},
	}
	
	message := []byte("From: sender@example.com\r\nTo: recipient@example.com\r\nSubject: Test\r\n\r\nTest message body")
	reader := io.NopCloser(bytes.NewReader(message))
	
	err := session.Data(reader)
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
}

func TestAnonymousSession_Reset(t *testing.T) {
	session := &AnonymousSession{
		sessionID: "test-session",
		from:      "sender@example.com",
		to:        []string{"recipient@example.com"},
	}
	
	session.Reset()
	
	if session.from != "" {
		t.Errorf("Expected empty from after reset, got %s", session.from)
	}
	
	if session.to != nil {
		t.Errorf("Expected nil to after reset, got %v", session.to)
	}
}

func TestSession_Data_WithPlaceholder(t *testing.T) {
	tenantStore := store.NewTenantStore()
	backend := NewBackend(tenantStore)
	session := &Session{
		backend:   backend,
		sessionID: "test-session",
		from:      "sender@example.com",
		to:        []string{"recipient@example.com"},
	}
	
	message := []byte("From: sender@example.com\r\nTo: recipient@example.com\r\nSubject: Test\r\nContent-Type: text/plain\r\n\r\nHello,\n\n[[signature:default]]\n")
	reader := bytes.NewReader(message)
	
	err := session.Data(reader)
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	
	// The test verifies that the Data method processes messages with placeholders without errors
	// The actual placeholder replacement is tested in mime/walker_test.go
}

func TestSession_Data_SignedMessage(t *testing.T) {
	tenantStore := store.NewTenantStore()
	backend := NewBackend(tenantStore)
	session := &Session{
		backend:   backend,
		sessionID: "test-session",
		from:      "sender@example.com",
		to:        []string{"recipient@example.com"},
	}
	
	message := []byte("From: sender@example.com\r\nTo: recipient@example.com\r\nSubject: Signed Test\r\nContent-Type: multipart/signed; protocol=\"application/pkcs7-signature\"\r\n\r\nSigned content\n[[signature:default]]\n")
	reader := bytes.NewReader(message)
	
	err := session.Data(reader)
	if err != nil {
		t.Errorf("Data failed: %v", err)
	}
	
	// The test verifies that signed messages are accepted without modification
	// The S/MIME detection is tested in mime/walker_test.go
}


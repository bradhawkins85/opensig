package relay

import (
	"io"
	"log"
	"strings"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/google/uuid"
	"github.com/your-org/opensig/server/internal/mime"
	"github.com/your-org/opensig/server/internal/store"
)

// Backend implements smtp.Backend interface
type Backend struct {
	tenantStore *store.TenantStore
	mimeWalker  *mime.Walker
}

// NewBackend creates a new SMTP backend
func NewBackend(tenantStore *store.TenantStore) *Backend {
	// Create a signature renderer function
	signatureRenderer := func(name string) (html string, text string, err error) {
		// For now, return a simple default signature
		// TODO: In M4, integrate with template store and rules engine
		html = `<div class="signature">
<p>Best regards,<br>
<strong>Default Signature</strong></p>
</div>`
		text = `Best regards,
Default Signature`
		return html, text, nil
	}
	
	return &Backend{
		tenantStore: tenantStore,
		mimeWalker:  mime.NewWalker(signatureRenderer),
	}
}

// NewSession creates a new SMTP session
func (b *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	sessionID := uuid.New().String()
	log.Printf("[%s] New SMTP connection from %s", sessionID, c.Conn().RemoteAddr())
	
	return &Session{
		backend:   b,
		sessionID: sessionID,
		conn:      c,
	}, nil
}

// Session implements smtp.Session interface
type Session struct {
	backend   *Backend
	sessionID string
	conn      *smtp.Conn
	from      string
	to        []string
	startTime time.Time
}

// AuthPlain implements PLAIN authentication
func (s *Session) AuthPlain(username, password string) error {
	log.Printf("[%s] AUTH PLAIN attempt for user: %s", s.sessionID, username)
	
	// Per-tenant authentication
	// Extract domain from username (e.g., user@example.com -> example.com)
	parts := strings.Split(username, "@")
	if len(parts) != 2 {
		log.Printf("[%s] AUTH PLAIN failed: invalid username format", s.sessionID)
		return smtp.ErrAuthFailed
	}
	
	domain := parts[1]
	
	// Verify tenant exists and is active
	tenants := s.backend.tenantStore.List()
	tenantFound := false
	for _, tenant := range tenants {
		if tenant.Domain == domain && tenant.Active {
			tenantFound = true
			break
		}
	}
	
	if !tenantFound {
		log.Printf("[%s] AUTH PLAIN failed: tenant not found or inactive for domain %s", s.sessionID, domain)
		return smtp.ErrAuthFailed
	}
	
	// In a no-op pass-through relay, we accept any password for now
	// TODO: In M4, implement proper credential validation
	log.Printf("[%s] AUTH PLAIN successful for user: %s (tenant: %s)", s.sessionID, username, domain)
	return nil
}

// Mail implements the MAIL FROM command
func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	s.startTime = time.Now()
	s.from = from
	log.Printf("[%s] MAIL FROM: <%s>", s.sessionID, from)
	return nil
}

// Rcpt implements the RCPT TO command
func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.to = append(s.to, to)
	log.Printf("[%s] RCPT TO: <%s>", s.sessionID, to)
	return nil
}

// Data implements the DATA command
func (s *Session) Data(r io.Reader) error {
	// Read the message
	data, err := io.ReadAll(r)
	if err != nil {
		log.Printf("[%s] Error reading message data: %v", s.sessionID, err)
		return err
	}
	
	messageSize := len(data)
	duration := time.Since(s.startTime)
	
	// Process the message with MIME walker
	modified, wasModified, err := s.backend.mimeWalker.ProcessMessage(data)
	if err != nil {
		log.Printf("[%s] Error processing message: %v", s.sessionID, err)
		// Continue with original message on error
		modified = data
		wasModified = false
	}
	
	modificationStatus := "unmodified"
	if wasModified {
		modificationStatus = "modified (placeholders replaced)"
	}
	
	log.Printf("[%s] Message received: from=<%s> to=%v size=%d bytes duration=%v status=%s", 
		s.sessionID, s.from, s.to, messageSize, duration, modificationStatus)
	
	// TODO: In M4, forward the modified message to the next hop
	// For now, we just process and log
	_ = modified
	
	return nil
}

// Reset resets the session state
func (s *Session) Reset() {
	log.Printf("[%s] RSET command", s.sessionID)
	s.from = ""
	s.to = nil
}

// Logout closes the session
func (s *Session) Logout() error {
	log.Printf("[%s] Session closed", s.sessionID)
	return nil
}

// AnonymousBackend implements smtp.Backend for anonymous (no auth) access
type AnonymousBackend struct {
	backend *Backend
}

// NewAnonymousBackend wraps a Backend to allow anonymous access
func NewAnonymousBackend(backend *Backend) *AnonymousBackend {
	return &AnonymousBackend{backend: backend}
}

// NewSession creates a new anonymous session
func (ab *AnonymousBackend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	sessionID := uuid.New().String()
	log.Printf("[%s] New anonymous SMTP connection from %s", sessionID, c.Conn().RemoteAddr())
	
	return &AnonymousSession{
		backend:   ab.backend,
		sessionID: sessionID,
	}, nil
}

// AnonymousSession implements smtp.Session for anonymous connections
type AnonymousSession struct {
	backend   *Backend
	sessionID string
	from      string
	to        []string
	startTime time.Time
}

// Mail implements the MAIL FROM command
func (s *AnonymousSession) Mail(from string, opts *smtp.MailOptions) error {
	s.startTime = time.Now()
	s.from = from
	log.Printf("[%s] MAIL FROM: <%s> (anonymous)", s.sessionID, from)
	return nil
}

// Rcpt implements the RCPT TO command
func (s *AnonymousSession) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.to = append(s.to, to)
	log.Printf("[%s] RCPT TO: <%s> (anonymous)", s.sessionID, to)
	return nil
}

// Data implements the DATA command
func (s *AnonymousSession) Data(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		log.Printf("[%s] Error reading message data: %v", s.sessionID, err)
		return err
	}
	
	messageSize := len(data)
	duration := time.Since(s.startTime)
	
	// Process the message with MIME walker
	modified, wasModified, err := s.backend.mimeWalker.ProcessMessage(data)
	if err != nil {
		log.Printf("[%s] Error processing message: %v", s.sessionID, err)
		// Continue with original message on error
		modified = data
		wasModified = false
	}
	
	modificationStatus := "unmodified"
	if wasModified {
		modificationStatus = "modified (placeholders replaced)"
	}
	
	log.Printf("[%s] Message received (anonymous): from=<%s> to=%v size=%d bytes duration=%v status=%s", 
		s.sessionID, s.from, s.to, messageSize, duration, modificationStatus)
	
	// TODO: In M4, forward the modified message to the next hop
	_ = modified
	
	return nil
}

// Reset resets the session state
func (s *AnonymousSession) Reset() {
	log.Printf("[%s] RSET command (anonymous)", s.sessionID)
	s.from = ""
	s.to = nil
}

// Logout closes the session
func (s *AnonymousSession) Logout() error {
	log.Printf("[%s] Session closed (anonymous)", s.sessionID)
	return nil
}

// Compile-time interface checks
var (
	_ smtp.Backend = (*Backend)(nil)
	_ smtp.Session = (*Session)(nil)
	_ smtp.Backend = (*AnonymousBackend)(nil)
	_ smtp.Session = (*AnonymousSession)(nil)
)

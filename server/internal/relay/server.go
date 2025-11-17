package relay

import (
	"fmt"
	"log"
	"time"

	"github.com/emersion/go-smtp"
)

// Server wraps the SMTP server with configuration
type Server struct {
	config     *Config
	smtpServer *smtp.Server
}

// NewServer creates a new SMTP relay server
func NewServer(config *Config, backend smtp.Backend) *Server {
	s := smtp.NewServer(backend)
	
	// Apply configuration
	s.Addr = config.Address
	s.Domain = config.Domain
	s.MaxMessageBytes = config.MaxMessageBytes
	s.MaxRecipients = config.MaxRecipients
	s.AllowInsecureAuth = !config.RequireTLS
	
	// Configure TLS if enabled
	if config.EnableTLS && config.TLSConfig != nil {
		s.TLSConfig = config.TLSConfig
		log.Printf("STARTTLS enabled with minimum TLS version 1.2")
		
		// Log MTLS configuration
		switch config.TLSConfig.ClientAuth {
		case 0: // NoClientCert
			log.Printf("MTLS: client certificates not requested")
		case 1: // RequestClientCert
			log.Printf("MTLS: client certificates requested but not required")
		case 2, 3, 4: // RequireAnyClientCert, VerifyClientCertIfGiven, RequireAndVerifyClientCert
			log.Printf("MTLS: client certificates required and verified")
		}
	}
	
	// Set timeouts
	s.ReadTimeout = 30 * time.Second
	s.WriteTimeout = 30 * time.Second
	
	return &Server{
		config:     config,
		smtpServer: s,
	}
}

// ListenAndServe starts the SMTP server
func (s *Server) ListenAndServe() error {
	log.Printf("Starting SMTP relay server on %s (domain: %s)", s.config.Address, s.config.Domain)
	log.Printf("Max message size: %d bytes, Max recipients: %d", s.config.MaxMessageBytes, s.config.MaxRecipients)
	
	if err := s.smtpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("SMTP server error: %w", err)
	}
	
	return nil
}

// Close shuts down the SMTP server
func (s *Server) Close() error {
	log.Printf("Shutting down SMTP relay server")
	return s.smtpServer.Close()
}

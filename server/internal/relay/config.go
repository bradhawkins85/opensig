package relay

import (
	"crypto/tls"
	"fmt"
	"os"
)

// Config holds the SMTP relay server configuration
type Config struct {
	// Address is the listen address (e.g., ":2525")
	Address string

	// Domain is the server's domain name for SMTP greeting
	Domain string

	// EnableTLS enables STARTTLS support
	EnableTLS bool

	// RequireTLS requires clients to use STARTTLS before authentication
	RequireTLS bool

	// TLSConfig is the TLS configuration for STARTTLS/MTLS
	TLSConfig *tls.Config

	// MaxMessageBytes is the maximum message size in bytes (default 10MB)
	MaxMessageBytes int64

	// MaxRecipients is the maximum number of recipients per message
	MaxRecipients int
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	domain := os.Getenv("SMTP_DOMAIN")
	if domain == "" {
		domain = "opensig.local"
	}

	return &Config{
		Address:         ":2525",
		Domain:          domain,
		EnableTLS:       true,
		RequireTLS:      false,
		MaxMessageBytes: 10 * 1024 * 1024, // 10MB
		MaxRecipients:   100,
	}
}

// LoadTLSConfig loads TLS configuration from certificate files
// If certFile and keyFile are empty, returns nil (TLS disabled)
func LoadTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	if certFile == "" || keyFile == "" {
		return nil, nil
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		ClientAuth:   tls.RequestClientCert, // Request but don't require client cert (for MTLS)
	}, nil
}

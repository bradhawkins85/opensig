package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/your-org/opensig/server/internal/relay"
	"github.com/your-org/opensig/server/internal/store"
)

func main() {
	log.Printf("OpenSig Relay (SMTP smart host) - M3: MTLS listener & message ingest")
	
	// Initialize tenant store (in-memory for now)
	tenantStore := store.NewTenantStore()
	
	// Create SMTP backend
	backend := relay.NewBackend(tenantStore)
	
	// For now, use anonymous backend to allow connections without auth
	// This makes it easier to test as a smart host
	anonBackend := relay.NewAnonymousBackend(backend)
	
	// Load configuration
	config := relay.DefaultConfig()
	
	// Load TLS configuration from environment if provided
	certFile := os.Getenv("SMTP_TLS_CERT")
	keyFile := os.Getenv("SMTP_TLS_KEY")
	
	if certFile != "" && keyFile != "" {
		tlsConfig, err := relay.LoadTLSConfig(certFile, keyFile)
		if err != nil {
			log.Printf("Warning: Failed to load TLS config: %v", err)
			log.Printf("Continuing without TLS support")
		} else {
			config.TLSConfig = tlsConfig
			log.Printf("TLS configuration loaded from %s", certFile)
		}
	} else {
		log.Printf("No TLS certificate configured (set SMTP_TLS_CERT and SMTP_TLS_KEY)")
		log.Printf("Running without STARTTLS support")
		config.EnableTLS = false
	}
	
	// Create and start server
	server := relay.NewServer(config, anonBackend)
	
	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			errChan <- err
		}
	}()
	
	// Wait for shutdown signal or error
	select {
	case err := <-errChan:
		log.Fatalf("Server error: %v", err)
	case sig := <-sigChan:
		log.Printf("Received signal %v, shutting down...", sig)
		if err := server.Close(); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	}
}

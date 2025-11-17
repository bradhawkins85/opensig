package main

import (
	"log"
	"net"
)

// This is a placeholder for the SMTP relay/smart host.
// In production, implement MTLS, MIME parsing and placeholder replacement here.
func main() {
	log.Printf("OpenSig Relay (SMTP smart host) – stub")
	// Reserve port or print config; implementation pending.
	ln, err := net.Listen("tcp", ":2525")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	log.Printf("Listening on :2525 (no-op relay stub)")
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept err: %v", err)
			continue
		}
		_ = conn.Close()
	}
}

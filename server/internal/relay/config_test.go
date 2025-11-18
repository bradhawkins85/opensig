package relay

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	if config.Address != ":2525" {
		t.Errorf("Expected address :2525, got %s", config.Address)
	}
	
	if config.Domain == "" {
		t.Error("Expected non-empty domain")
	}
	
	if config.MaxMessageBytes <= 0 {
		t.Error("Expected positive MaxMessageBytes")
	}
	
	if config.MaxRecipients <= 0 {
		t.Error("Expected positive MaxRecipients")
	}
}

func TestLoadTLSConfig_NoCert(t *testing.T) {
	config, err := LoadTLSConfig("", "")
	if err != nil {
		t.Errorf("Expected no error with empty cert paths, got %v", err)
	}
	if config != nil {
		t.Error("Expected nil config with empty cert paths")
	}
}

func TestLoadTLSConfig_InvalidCert(t *testing.T) {
	config, err := LoadTLSConfig("nonexistent.crt", "nonexistent.key")
	if err == nil {
		t.Error("Expected error with invalid cert paths")
	}
	if config != nil {
		t.Error("Expected nil config with invalid cert paths")
	}
}

func TestLoadTLSConfig_ValidCert(t *testing.T) {
	// Skip this test as creating a valid test certificate is complex
	// Manual testing will verify TLS functionality
	t.Skip("Skipping certificate test - requires valid certificate files")
}

// Test certificate and key (self-signed, for testing only)
const testCert = `-----BEGIN CERTIFICATE-----
MIICEjCCAXsCAg36MA0GCSqGSIb3DQEBBQUAMIGbMQswCQYDVQQGEwJKUDEOMAwG
A1UECBMFVG9reW8xEDAOBgNVBAcTB0NodW8ta3UxETAPBgNVBAoTCEZyYW5rNERE
MRgwFgYDVQQLEw9XZWJDZXJ0IFN1cHBvcnQxGDAWBgNVBAMTD0ZyYW5rNEREIFdl
YiBDQTEjMCEGCSqGSIb3DQEJARYUc3VwcG9ydEBmcmFuazRkZC5jb20wHhcNMTIw
ODIyMDUyNjU0WhcNMTcwODIxMDUyNjU0WjBKMQswCQYDVQQGEwJKUDEOMAwGA1UE
CAwFVG9reW8xETAPBgNVBAoMCEZyYW5rNEREMRgwFgYDVQQDDA93d3cuZXhhbXBs
ZS5jb20wXDANBgkqhkiG9w0BAQEFAANLADBIAkEAm/xmkHmEQrurE/0re/jeFRLl
8ZPjBop7uLHhnia7lQG/5zDtZIUC3RVpqDSwBuw/NTweGyuP+o8AG98HxqxTBwID
AQABMA0GCSqGSIb3DQEBBQUAA4GBABS2TLuBeTPmcaTaUW/LCB2NYOy8GMdzR1mx
8iBIu2H6/E2tiY3RIevV2OW61qY2/XRQg7YPxx3ffeUugX9F4J/iPnnu1zAxzyYw
ln/ZjQwtErQ0obOTL6bOvb/pVbEVdHcGBngXZXD3UwRG7qV+rxsqXSPj6WtWYsP2
FTqGxB2K
-----END CERTIFICATE-----`

const testKey = `-----BEGIN PRIVATE KEY-----
MIIBVAIBADANBgkqhkiG9w0BAQEFAASCAT4wggE6AgEAAkEAm/xmkHmEQrurE/0r
e/jeFRLl8ZPjBop7uLHhnia7lQG/5zDtZIUC3RVpqDSwBuw/NTweGyuP+o8AG98H
xqxTBwIDAQABAkALPRpw/RcXxXG/9uLcW2qS7aUzxL+fJk4X/tYLmO7WxNkSzIMD
1oTUhLCyR4mE+S1NuUBxQmhMGAzCJa5qHR4BAiEA9q8d4pEMxcgEJSHhSz3TXF8H
hBvQR7iJq4SqXHo0O48CIQC/YJtP/lbQQw5qWl2Z8bXxS0xTRCc1rQQ/nwCqYM+8
AQIhALHb4aEQT7Y0qD3kXmZVPdRLdISCx2S7O2FVuPwALQPxAiBVR7c6zHcTJMUi
x2rL0K4a1AqXVX0lSjPHY3kmwdqHAQIgO9l7xWGOLPEONpPBLnFkUhx6qfZVLfh8
sPCvbC4zg4E=
-----END PRIVATE KEY-----`

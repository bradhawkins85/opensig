package mime

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// mockSignatureRenderer is a test renderer that returns simple signatures
func mockSignatureRenderer(name string) (html string, text string, err error) {
	switch name {
	case "default":
		html = `<div class="signature">
<p>Best regards,<br>
<strong>John Doe</strong><br>
Software Engineer</p>
</div>`
		text = `Best regards,
John Doe
Software Engineer`
	case "marketing":
		html = `<div class="signature">
<p><strong>SPECIAL OFFER!</strong><br>
Contact us today!</p>
</div>`
		text = `SPECIAL OFFER!
Contact us today!`
	case "professional":
		html = `<div class="signature">
<p>Sincerely,<br>
<strong>Jane Smith</strong></p>
</div>`
		text = `Sincerely,
Jane Smith`
	default:
		html = ""
		text = ""
	}
	return html, text, nil
}

func TestWalker_ParseMessage_SimpleText(t *testing.T) {
	data, err := os.ReadFile("testdata/simple_text.eml")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	
	walker := NewWalker(mockSignatureRenderer)
	info, err := walker.ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}
	
	if info.IsSigned {
		t.Error("Expected message not to be signed")
	}
	if info.IsEncrypted {
		t.Error("Expected message not to be encrypted")
	}
	if !info.HasPlaceholders {
		t.Error("Expected message to have placeholders")
	}
}

func TestWalker_ParseMessage_SimpleHTML(t *testing.T) {
	data, err := os.ReadFile("testdata/simple_html.eml")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	
	walker := NewWalker(mockSignatureRenderer)
	info, err := walker.ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}
	
	if info.IsSigned {
		t.Error("Expected message not to be signed")
	}
	if info.IsEncrypted {
		t.Error("Expected message not to be encrypted")
	}
	if !info.HasPlaceholders {
		t.Error("Expected message to have placeholders")
	}
}

func TestWalker_ParseMessage_Multipart(t *testing.T) {
	data, err := os.ReadFile("testdata/multipart.eml")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	
	walker := NewWalker(mockSignatureRenderer)
	info, err := walker.ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}
	
	if info.IsSigned {
		t.Error("Expected message not to be signed")
	}
	if info.IsEncrypted {
		t.Error("Expected message not to be encrypted")
	}
	if !info.HasPlaceholders {
		t.Error("Expected message to have placeholders")
	}
}

func TestWalker_ParseMessage_Signed(t *testing.T) {
	data, err := os.ReadFile("testdata/signed.eml")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	
	walker := NewWalker(mockSignatureRenderer)
	info, err := walker.ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}
	
	if !info.IsSigned {
		t.Error("Expected message to be signed")
	}
	if info.IsEncrypted {
		t.Error("Expected message not to be encrypted")
	}
}

func TestWalker_ParseMessage_Encrypted(t *testing.T) {
	data, err := os.ReadFile("testdata/encrypted.eml")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	
	walker := NewWalker(mockSignatureRenderer)
	info, err := walker.ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}
	
	if info.IsSigned {
		t.Error("Expected message not to be signed")
	}
	if !info.IsEncrypted {
		t.Error("Expected message to be encrypted")
	}
}

func TestWalker_ParseMessage_NoPlaceholder(t *testing.T) {
	data, err := os.ReadFile("testdata/no_placeholder.eml")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	
	walker := NewWalker(mockSignatureRenderer)
	info, err := walker.ParseMessage(data)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}
	
	if info.IsSigned {
		t.Error("Expected message not to be signed")
	}
	if info.IsEncrypted {
		t.Error("Expected message not to be encrypted")
	}
	if info.HasPlaceholders {
		t.Error("Expected message not to have placeholders")
	}
}

func TestWalker_ProcessMessage_SimpleText(t *testing.T) {
	data, err := os.ReadFile("testdata/simple_text.eml")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	
	walker := NewWalker(mockSignatureRenderer)
	modified, wasModified, err := walker.ProcessMessage(data)
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}
	
	if !wasModified {
		t.Error("Expected message to be modified")
	}
	
	modifiedStr := string(modified)
	
	// Should not contain the placeholder
	if strings.Contains(modifiedStr, "[[signature:default]]") {
		t.Error("Expected placeholder to be replaced")
	}
	
	// Should contain the signature text
	if !strings.Contains(modifiedStr, "John Doe") {
		t.Error("Expected signature content to be present")
	}
	if !strings.Contains(modifiedStr, "Software Engineer") {
		t.Error("Expected signature content to be present")
	}
}

func TestWalker_ProcessMessage_SimpleHTML(t *testing.T) {
	data, err := os.ReadFile("testdata/simple_html.eml")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	
	walker := NewWalker(mockSignatureRenderer)
	modified, wasModified, err := walker.ProcessMessage(data)
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}
	
	if !wasModified {
		t.Error("Expected message to be modified")
	}
	
	modifiedStr := string(modified)
	
	// Should not contain the placeholder
	if strings.Contains(modifiedStr, "[[signature:marketing]]") {
		t.Error("Expected placeholder to be replaced")
	}
	
	// Should contain the HTML signature
	if !strings.Contains(modifiedStr, "SPECIAL OFFER!") {
		t.Error("Expected signature content to be present")
	}
	if !strings.Contains(modifiedStr, "Contact us today!") {
		t.Error("Expected signature content to be present")
	}
}

func TestWalker_ProcessMessage_Multipart(t *testing.T) {
	data, err := os.ReadFile("testdata/multipart.eml")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	
	walker := NewWalker(mockSignatureRenderer)
	modified, wasModified, err := walker.ProcessMessage(data)
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}
	
	if !wasModified {
		t.Error("Expected message to be modified")
	}
	
	modifiedStr := string(modified)
	
	// Should not contain the placeholder
	if strings.Contains(modifiedStr, "[[signature:default]]") {
		t.Error("Expected placeholder to be replaced")
	}
	
	// Should contain both text and HTML signatures
	if !strings.Contains(modifiedStr, "John Doe") {
		t.Error("Expected signature content to be present in text part")
	}
	if !strings.Contains(modifiedStr, "Software Engineer") {
		t.Error("Expected signature content to be present")
	}
}

func TestWalker_ProcessMessage_Signed_Untouched(t *testing.T) {
	data, err := os.ReadFile("testdata/signed.eml")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	
	walker := NewWalker(mockSignatureRenderer)
	modified, wasModified, err := walker.ProcessMessage(data)
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}
	
	if wasModified {
		t.Error("Expected signed message not to be modified")
	}
	
	// Should be identical to original
	if !bytes.Equal(data, modified) {
		t.Error("Expected signed message to remain unchanged")
	}
}

func TestWalker_ProcessMessage_Encrypted_Untouched(t *testing.T) {
	data, err := os.ReadFile("testdata/encrypted.eml")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	
	walker := NewWalker(mockSignatureRenderer)
	modified, wasModified, err := walker.ProcessMessage(data)
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}
	
	if wasModified {
		t.Error("Expected encrypted message not to be modified")
	}
	
	// Should be identical to original
	if !bytes.Equal(data, modified) {
		t.Error("Expected encrypted message to remain unchanged")
	}
}

func TestWalker_ProcessMessage_NoPlaceholder_Untouched(t *testing.T) {
	data, err := os.ReadFile("testdata/no_placeholder.eml")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	
	walker := NewWalker(mockSignatureRenderer)
	modified, wasModified, err := walker.ProcessMessage(data)
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}
	
	if wasModified {
		t.Error("Expected message without placeholders not to be modified")
	}
	
	// Should be identical to original
	if !bytes.Equal(data, modified) {
		t.Error("Expected message without placeholders to remain unchanged")
	}
}

func TestHasPlaceholder(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		expected bool
	}{
		{
			name:     "with default placeholder",
			data:     "Hello [[signature:default]] world",
			expected: true,
		},
		{
			name:     "with marketing placeholder",
			data:     "Test [[signature:marketing]]",
			expected: true,
		},
		{
			name:     "without placeholder",
			data:     "Hello world",
			expected: false,
		},
		{
			name:     "with malformed bracket",
			data:     "Hello [signature:default] world",
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasPlaceholder([]byte(tt.data))
			if result != tt.expected {
				t.Errorf("HasPlaceholder() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractPlaceholders(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		expected []string
	}{
		{
			name:     "single placeholder",
			data:     "Hello [[signature:default]] world",
			expected: []string{"default"},
		},
		{
			name:     "multiple placeholders",
			data:     "[[signature:default]] and [[signature:marketing]]",
			expected: []string{"default", "marketing"},
		},
		{
			name:     "no placeholders",
			data:     "Hello world",
			expected: nil,
		},
		{
			name:     "duplicate placeholders",
			data:     "[[signature:default]] and [[signature:default]]",
			expected: []string{"default", "default"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractPlaceholders([]byte(tt.data))
			if len(result) != len(tt.expected) {
				t.Errorf("ExtractPlaceholders() returned %d items, want %d", len(result), len(tt.expected))
				return
			}
			for i, name := range result {
				if name != tt.expected[i] {
					t.Errorf("ExtractPlaceholders()[%d] = %v, want %v", i, name, tt.expected[i])
				}
			}
		})
	}
}

func TestWalker_isSMIMESigned(t *testing.T) {
	walker := NewWalker(nil)
	
	tests := []struct {
		name        string
		contentType string
		expected    bool
	}{
		{
			name:        "multipart/signed",
			contentType: "multipart/signed; protocol=\"application/pkcs7-signature\"",
			expected:    true,
		},
		{
			name:        "application/pkcs7-signature",
			contentType: "application/pkcs7-signature",
			expected:    true,
		},
		{
			name:        "application/x-pkcs7-signature",
			contentType: "application/x-pkcs7-signature",
			expected:    true,
		},
		{
			name:        "text/plain",
			contentType: "text/plain",
			expected:    false,
		},
		{
			name:        "text/html",
			contentType: "text/html",
			expected:    false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := walker.isSMIMESigned(tt.contentType)
			if result != tt.expected {
				t.Errorf("isSMIMESigned(%s) = %v, want %v", tt.contentType, result, tt.expected)
			}
		})
	}
}

func TestWalker_isSMIMEEncrypted(t *testing.T) {
	walker := NewWalker(nil)
	
	tests := []struct {
		name        string
		contentType string
		expected    bool
	}{
		{
			name:        "application/pkcs7-mime",
			contentType: "application/pkcs7-mime; smime-type=enveloped-data",
			expected:    true,
		},
		{
			name:        "application/x-pkcs7-mime",
			contentType: "application/x-pkcs7-mime; smime-type=enveloped-data",
			expected:    true,
		},
		{
			name:        "text/plain",
			contentType: "text/plain",
			expected:    false,
		},
		{
			name:        "text/html",
			contentType: "text/html",
			expected:    false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := walker.isSMIMEEncrypted(tt.contentType)
			if result != tt.expected {
				t.Errorf("isSMIMEEncrypted(%s) = %v, want %v", tt.contentType, result, tt.expected)
			}
		})
	}
}

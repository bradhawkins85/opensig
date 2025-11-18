package mime

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"regexp"
	"strings"
)

var (
	// signaturePlaceholderRegex matches [[signature:name]] patterns
	signaturePlaceholderRegex = regexp.MustCompile(`\[\[signature:([^\]]+)\]\]`)
)

// MessageInfo contains parsed information about an email message
type MessageInfo struct {
	// Original message data
	RawMessage []byte
	
	// Parsed message
	Message *mail.Message
	
	// Whether the message is S/MIME signed or encrypted
	IsSigned    bool
	IsEncrypted bool
	
	// Whether the message contains signature placeholders
	HasPlaceholders bool
	
	// The modified message (if placeholders were replaced)
	ModifiedMessage []byte
}

// Walker provides methods for walking MIME message structures
type Walker struct {
	// SignatureRenderer is a function that renders a signature given its name
	SignatureRenderer func(name string) (html string, text string, err error)
}

// NewWalker creates a new MIME walker
func NewWalker(renderer func(string) (string, string, error)) *Walker {
	return &Walker{
		SignatureRenderer: renderer,
	}
}

// ParseMessage parses an email message and returns information about it
func (w *Walker) ParseMessage(data []byte) (*MessageInfo, error) {
	info := &MessageInfo{
		RawMessage: data,
	}
	
	// Parse the message
	msg, err := mail.ReadMessage(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}
	info.Message = msg
	
	// Check for S/MIME signatures and encryption
	contentType := msg.Header.Get("Content-Type")
	info.IsSigned = w.isSMIMESigned(contentType)
	info.IsEncrypted = w.isSMIMEEncrypted(contentType)
	
	// If the message is signed or encrypted, don't process it
	if info.IsSigned || info.IsEncrypted {
		return info, nil
	}
	
	// Check if message contains placeholders
	body, err := io.ReadAll(msg.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read message body: %w", err)
	}
	
	// Check for placeholders in the raw body
	info.HasPlaceholders = signaturePlaceholderRegex.MatchString(string(body))
	
	return info, nil
}

// ProcessMessage processes a message, replacing signature placeholders
// Returns the modified message or the original if no changes were needed
func (w *Walker) ProcessMessage(data []byte) ([]byte, bool, error) {
	info, err := w.ParseMessage(data)
	if err != nil {
		return nil, false, err
	}
	
	// Skip signed or encrypted messages
	if info.IsSigned || info.IsEncrypted {
		return data, false, nil
	}
	
	// If no placeholders, return original
	if !info.HasPlaceholders {
		return data, false, nil
	}
	
	// Process the message and replace placeholders
	modified, err := w.replacePlaceholders(data, info)
	if err != nil {
		return nil, false, err
	}
	
	return modified, true, nil
}

// replacePlaceholders replaces signature placeholders in the message
func (w *Walker) replacePlaceholders(data []byte, info *MessageInfo) ([]byte, error) {
	msg := info.Message
	contentType := msg.Header.Get("Content-Type")
	
	// Parse content type
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		// If no content type, treat as plain text
		mediaType = "text/plain"
	}
	
	// Split headers and body
	headerEnd := bytes.Index(data, []byte("\r\n\r\n"))
	if headerEnd == -1 {
		headerEnd = bytes.Index(data, []byte("\n\n"))
		if headerEnd == -1 {
			return nil, fmt.Errorf("failed to find header/body boundary")
		}
	}
	
	headers := data[:headerEnd]
	bodyStart := headerEnd + len([]byte("\r\n\r\n"))
	if bytes.Contains(data[headerEnd:headerEnd+4], []byte("\n\n")) {
		bodyStart = headerEnd + len([]byte("\n\n"))
	}
	body := data[bodyStart:]
	
	var newBody []byte
	
	if strings.HasPrefix(mediaType, "multipart/") {
		// Handle multipart messages
		boundary := params["boundary"]
		if boundary == "" {
			return nil, fmt.Errorf("multipart message missing boundary")
		}
		
		newBody, err = w.processMultipart(body, boundary)
		if err != nil {
			return nil, err
		}
	} else {
		// Handle single-part messages
		newBody, err = w.processSinglePart(body, mediaType)
		if err != nil {
			return nil, err
		}
	}
	
	// Reconstruct the message
	result := bytes.NewBuffer(make([]byte, 0, len(headers)+len(newBody)+4))
	result.Write(headers)
	result.WriteString("\r\n\r\n")
	result.Write(newBody)
	
	return result.Bytes(), nil
}

// processMultipart processes a multipart message body
func (w *Walker) processMultipart(body []byte, boundary string) ([]byte, error) {
	reader := multipart.NewReader(bytes.NewReader(body), boundary)
	
	var parts [][]byte
	
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read multipart: %w", err)
		}
		
		partData, err := io.ReadAll(part)
		if err != nil {
			return nil, fmt.Errorf("failed to read part data: %w", err)
		}
		
		contentType := part.Header.Get("Content-Type")
		mediaType, _, _ := mime.ParseMediaType(contentType)
		
		// Only process text parts
		if strings.HasPrefix(mediaType, "text/") {
			processed, err := w.processSinglePart(partData, mediaType)
			if err != nil {
				return nil, err
			}
			
			// Reconstruct the part with headers
			var buf bytes.Buffer
			buf.WriteString("--" + boundary + "\r\n")
			for key, values := range part.Header {
				for _, value := range values {
					buf.WriteString(key + ": " + value + "\r\n")
				}
			}
			buf.WriteString("\r\n")
			buf.Write(processed)
			parts = append(parts, buf.Bytes())
		} else {
			// Keep non-text parts as-is
			var buf bytes.Buffer
			buf.WriteString("--" + boundary + "\r\n")
			for key, values := range part.Header {
				for _, value := range values {
					buf.WriteString(key + ": " + value + "\r\n")
				}
			}
			buf.WriteString("\r\n")
			buf.Write(partData)
			parts = append(parts, buf.Bytes())
		}
	}
	
	// Combine all parts
	var result bytes.Buffer
	for _, part := range parts {
		result.Write(part)
		result.WriteString("\r\n")
	}
	result.WriteString("--" + boundary + "--\r\n")
	
	return result.Bytes(), nil
}

// processSinglePart processes a single message part, replacing placeholders
func (w *Walker) processSinglePart(body []byte, mediaType string) ([]byte, error) {
	bodyStr := string(body)
	
	// Find all signature placeholders
	matches := signaturePlaceholderRegex.FindAllStringSubmatch(bodyStr, -1)
	if len(matches) == 0 {
		return body, nil
	}
	
	// Replace each placeholder
	result := bodyStr
	for _, match := range matches {
		placeholder := match[0] // Full match: [[signature:name]]
		signatureName := match[1] // Captured group: name
		
		// Get the signature content
		var replacement string
		if w.SignatureRenderer != nil {
			var html, text string
			var err error
			html, text, err = w.SignatureRenderer(signatureName)
			if err != nil {
				// If we can't render the signature, leave the placeholder
				continue
			}
			
			// Use HTML for text/html, text for text/plain
			if strings.Contains(mediaType, "html") {
				replacement = html
			} else {
				replacement = text
			}
		} else {
			// No renderer, remove placeholder
			replacement = ""
		}
		
		result = strings.ReplaceAll(result, placeholder, replacement)
	}
	
	return []byte(result), nil
}

// isSMIMESigned checks if the message is S/MIME signed
func (w *Walker) isSMIMESigned(contentType string) bool {
	mediaType, _, _ := mime.ParseMediaType(contentType)
	
	// S/MIME signed messages have specific content types
	return mediaType == "multipart/signed" ||
		mediaType == "application/pkcs7-signature" ||
		mediaType == "application/x-pkcs7-signature"
}

// isSMIMEEncrypted checks if the message is S/MIME encrypted
func (w *Walker) isSMIMEEncrypted(contentType string) bool {
	mediaType, _, _ := mime.ParseMediaType(contentType)
	
	// S/MIME encrypted messages have specific content types
	return mediaType == "application/pkcs7-mime" ||
		mediaType == "application/x-pkcs7-mime"
}

// HasPlaceholder checks if a message contains any signature placeholders
func HasPlaceholder(data []byte) bool {
	return signaturePlaceholderRegex.Match(data)
}

// ExtractPlaceholders extracts all signature placeholder names from a message
func ExtractPlaceholders(data []byte) []string {
	matches := signaturePlaceholderRegex.FindAllStringSubmatch(string(data), -1)
	var names []string
	for _, match := range matches {
		if len(match) > 1 {
			names = append(names, match[1])
		}
	}
	return names
}

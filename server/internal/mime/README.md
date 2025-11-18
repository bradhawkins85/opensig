# MIME Walker Module

The MIME walker module provides functionality for parsing email messages and replacing signature placeholders with rendered content.

## Features

- **Placeholder Detection**: Detects `[[signature:name]]` patterns in email bodies
- **Content Type Support**: Handles both HTML (`text/html`) and plain text (`text/plain`)
- **Multipart MIME**: Correctly processes multipart messages with multiple alternatives
- **S/MIME Protection**: Automatically skips signed and encrypted messages to preserve cryptographic integrity
- **Safe Processing**: Returns original message unchanged if no placeholders found or if processing encounters errors

## Usage

```go
import "github.com/your-org/opensig/server/internal/mime"

// Create a signature renderer function
renderer := func(name string) (html string, text string, err error) {
    // Fetch signature from database/template store
    // Return HTML and text versions
    return htmlSignature, textSignature, nil
}

// Create walker
walker := mime.NewWalker(renderer)

// Process a message
modified, wasModified, err := walker.ProcessMessage(emailData)
if err != nil {
    // Handle error
}

if wasModified {
    // Use modified message with signatures inserted
} else {
    // Use original message
}
```

## Placeholder Format

Placeholders should be in the format: `[[signature:name]]`

Examples:
- `[[signature:default]]` - Default signature
- `[[signature:marketing]]` - Marketing signature
- `[[signature:professional]]` - Professional signature

## S/MIME Detection

The walker automatically detects and skips the following message types:

### Signed Messages
- `multipart/signed`
- `application/pkcs7-signature`
- `application/x-pkcs7-signature`

### Encrypted Messages
- `application/pkcs7-mime`
- `application/x-pkcs7-mime`

These messages are returned unchanged to preserve their cryptographic integrity.

## Testing

The module includes comprehensive test coverage with EML fixtures:

- `simple_text.eml` - Plain text email with placeholder
- `simple_html.eml` - HTML email with placeholder
- `multipart.eml` - Multipart email (text + HTML) with placeholders
- `signed.eml` - S/MIME signed message (should be untouched)
- `encrypted.eml` - S/MIME encrypted message (should be untouched)
- `no_placeholder.eml` - Message without placeholders

Run tests with:
```bash
go test ./internal/mime/...
```

## Architecture

### Components

1. **Walker**: Main struct that orchestrates MIME parsing and placeholder replacement
2. **MessageInfo**: Struct containing parsed message metadata (signed, encrypted, has placeholders)
3. **Helper Functions**:
   - `HasPlaceholder()`: Quick check if message contains placeholders
   - `ExtractPlaceholders()`: Extract all placeholder names from a message

### Processing Flow

1. Parse email headers and body
2. Check Content-Type for S/MIME indicators
3. If signed/encrypted, return original unchanged
4. Scan body for placeholders
5. If no placeholders, return original unchanged
6. For single-part messages: replace placeholders directly
7. For multipart messages: walk each part and replace placeholders in text parts
8. Return modified message

## Integration with Relay

The MIME walker is integrated into the SMTP relay backend:

```go
// In backend.go
backend := &Backend{
    tenantStore: tenantStore,
    mimeWalker:  mime.NewWalker(signatureRenderer),
}

// In Data() method
modified, wasModified, err := s.backend.mimeWalker.ProcessMessage(data)
```

The relay logs modification status for observability:
```
Message received: from=<sender@example.com> to=[recipient@example.com] size=1234 bytes duration=50ms status=modified (placeholders replaced)
```

## Future Enhancements

- Support for nested multipart messages
- Configurable placeholder patterns
- Caching of rendered signatures
- Metrics for placeholder replacement rates
- Support for conditional signature insertion based on rules

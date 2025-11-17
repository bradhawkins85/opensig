# Authentication Package

This package provides Microsoft Graph authentication using OAuth 2.0 for OpenSig.

## Features

- **OAuth 2.0 Authorization Code Flow** - Web-based authentication
- **Device Code Flow** - For CLI tools and headless environments
- **Token Storage** - Pluggable token storage interface
- **Minimal Scopes** - User.Read and offline_access by default
- **State Validation** - CSRF protection for OAuth flows

## Usage

### Initialize the Auth Service

```go
import (
    "github.com/your-org/opensig/server/internal/auth"
)

// Load configuration from environment
config := auth.NewConfigFromEnv()

// Create token store (in-memory for dev, use database for production)
tokenStore := auth.NewInMemoryTokenStore()

// Create auth service
authService, err := auth.NewService(config, tokenStore)
if err != nil {
    log.Fatal(err)
}
```

### Web Flow (Authorization Code)

```go
// Generate auth URL
authURL, state, err := authService.GetAuthURL()
if err != nil {
    log.Fatal(err)
}

// Redirect user to authURL
// User logs in via Microsoft
// Microsoft redirects back to your callback URL with code and state

// In your callback handler:
if !authService.ValidateState(state) {
    // Invalid state - potential CSRF attack
    return
}

tokenData, err := authService.ExchangeCodeForToken(ctx, code)
if err != nil {
    log.Fatal(err)
}

// Token is automatically stored and can be retrieved later
fmt.Printf("User authenticated: %s\n", tokenData.Email)
```

### Device Code Flow

```go
// Initiate device code flow
tokenData, err := authService.DeviceCodeFlow(ctx)
if err != nil {
    log.Fatal(err)
}

// The method will print instructions for the user to follow
// After user completes authentication, token is returned
fmt.Printf("Authenticated: %s\n", tokenData.Email)
```

### Retrieve Stored Tokens

```go
tokenData, err := authService.GetToken(userID)
if err == auth.ErrTokenNotFound {
    // User not authenticated
    return
}
if err != nil {
    log.Fatal(err)
}

// Use the access token to call Microsoft Graph
// tokenData.AccessToken
```

### Call Microsoft Graph API

```go
req, _ := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me", nil)
req.Header.Set("Authorization", "Bearer "+tokenData.AccessToken)

resp, err := http.DefaultClient.Do(req)
// Handle response...
```

## Environment Variables

- `AZURE_CLIENT_ID` - **Required** - Application (client) ID from Azure Portal
- `AZURE_CLIENT_SECRET` - Required for web flow - Client secret from Azure Portal
- `AZURE_TENANT_ID` - Optional - Tenant ID (defaults to "common" for multi-tenant)
- `AZURE_REDIRECT_URL` - Optional - Redirect URI (defaults to http://localhost:8080/auth/callback)

## Token Storage

The package provides a `TokenStore` interface:

```go
type TokenStore interface {
    Store(userID string, token *TokenData) error
    Get(userID string) (*TokenData, error)
    Delete(userID string) error
}
```

### In-Memory Store (Development Only)

```go
tokenStore := auth.NewInMemoryTokenStore()
```

⚠️ **Warning**: The in-memory store loses all tokens when the server restarts. Not suitable for production!

### Production Token Storage

For production, implement the `TokenStore` interface with:
- Encrypted database storage
- Redis with encryption at rest
- Azure Key Vault or similar secret managers

Example:

```go
type DatabaseTokenStore struct {
    db *sql.DB
}

func (s *DatabaseTokenStore) Store(userID string, token *auth.TokenData) error {
    // Encrypt token before storing
    encrypted, err := encrypt(token)
    if err != nil {
        return err
    }
    
    // Store in database
    _, err = s.db.Exec(
        "INSERT INTO tokens (user_id, data) VALUES ($1, $2) ON CONFLICT (user_id) DO UPDATE SET data = $2",
        userID, encrypted,
    )
    return err
}

// Implement Get and Delete...
```

## Security Considerations

### Development vs Production

**Development** (current implementation):
- ✅ In-memory token storage
- ✅ Simple and fast for testing
- ❌ Tokens lost on restart
- ❌ Not suitable for production

**Production** (recommended):
- ✅ Encrypted database storage
- ✅ Token refresh handling
- ✅ Secure session cookies
- ✅ HTTPS only
- ✅ Regular secret rotation

### Best Practices

1. **Never commit secrets** - Use environment variables or secret managers
2. **Use HTTPS in production** - Never send tokens over HTTP
3. **Rotate secrets regularly** - Update client secrets every 6 months
4. **Implement token refresh** - Refresh expired tokens automatically
5. **Validate redirect URIs** - Only allow registered URIs
6. **Implement proper logging** - Audit authentication events

## Error Handling

```go
tokenData, err := authService.ExchangeCodeForToken(ctx, code)
if err != nil {
    // Check for specific errors
    if strings.Contains(err.Error(), "invalid_grant") {
        // Authorization code expired or invalid
        return
    }
    if strings.Contains(err.Error(), "invalid_client") {
        // Client ID or secret incorrect
        return
    }
    // Other error
    return
}
```

## Testing

Mock the `AuthService` interface for testing:

```go
type mockAuthService struct {
    tokenData *auth.TokenData
    err       error
}

func (m *mockAuthService) GetAuthURL() (string, string, error) {
    return "https://login.microsoft.com/...", "state123", m.err
}

func (m *mockAuthService) ExchangeCodeForToken(ctx context.Context, code string) (*auth.TokenData, error) {
    return m.tokenData, m.err
}

// Use in tests
mockService := &mockAuthService{
    tokenData: &auth.TokenData{
        UserID: "test-user",
        Email:  "test@example.com",
    },
}
```

## References

- [Microsoft Identity Platform](https://docs.microsoft.com/en-us/azure/active-directory/develop/)
- [MSAL for Go](https://github.com/AzureAD/microsoft-authentication-library-for-go)
- [Microsoft Graph API](https://docs.microsoft.com/en-us/graph/)

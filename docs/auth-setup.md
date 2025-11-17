# Microsoft Graph Authentication Setup

This guide walks you through setting up Microsoft Graph authentication (Entra ID) for OpenSig.

## Overview

OpenSig uses OAuth 2.0 to authenticate users with Microsoft Entra ID (formerly Azure AD) and obtain access tokens to call Microsoft Graph APIs. Two authentication flows are supported:

1. **Authorization Code Flow** - For web applications (requires client secret)
2. **Device Code Flow** - For local development and headless environments

## Prerequisites

- An Azure subscription or Microsoft 365 tenant
- Administrator access to register applications in Azure Portal

## Step 1: Register an Application in Azure Portal

1. Navigate to [Azure Portal - App Registrations](https://portal.azure.com/#view/Microsoft_AAD_RegisteredApps/ApplicationsListBlade)

2. Click **New registration**

3. Configure the application:
   - **Name**: `OpenSig` (or your preferred name)
   - **Supported account types**: Choose based on your needs:
     - "Accounts in this organizational directory only" - Single tenant
     - "Accounts in any organizational directory" - Multi-tenant
   - **Redirect URI**: 
     - Platform: `Web`
     - URI: `http://localhost:8080/auth/callback` (for local dev)
     - For production, use your actual domain: `https://yourdomain.com/auth/callback`

4. Click **Register**

## Step 2: Configure API Permissions

1. In your app registration, go to **API permissions**

2. Click **Add a permission** → **Microsoft Graph** → **Delegated permissions**

3. Add the following permissions:
   - `User.Read` - Read user profile information
   - `offline_access` - Maintain access to data you have given it access to

4. Click **Grant admin consent** (requires admin privileges)

## Step 3: Create a Client Secret (for Web Flow)

> **Note**: Client secrets are only needed for web-based authentication flow. Device code flow doesn't require secrets.

1. Go to **Certificates & secrets**

2. Click **New client secret**

3. Add a description (e.g., "OpenSig Dev Secret")

4. Choose an expiration period (recommend 6 months for dev, shorter for production)

5. Click **Add**

6. **Important**: Copy the secret value immediately - it won't be shown again!

## Step 4: Configure Application Settings

1. Go to **Authentication** in your app registration

2. Under **Implicit grant and hybrid flows**, enable:
   - ✓ ID tokens (for authentication flows)

3. Under **Advanced settings**:
   - Allow public client flows: **Yes** (enables device code flow)

## Step 5: Configure Environment Variables

Create a `.env` file or set environment variables:

```bash
# Required - Application (client) ID from Azure Portal
AZURE_CLIENT_ID=your-client-id-here

# Required for web flow - Client secret value from Azure Portal
AZURE_CLIENT_SECRET=your-client-secret-here

# Tenant ID (optional, defaults to "common" for multi-tenant)
# Use your specific tenant ID for single-tenant apps
AZURE_TENANT_ID=common

# Redirect URL (must match what's configured in Azure Portal)
AZURE_REDIRECT_URL=http://localhost:8080/auth/callback
```

## Step 6: Test Authentication

### Web Flow (Authorization Code)

1. Start the server:
   ```bash
   cd server
   go run cmd/opensig-api/main.go
   ```

2. Open your browser and navigate to:
   ```
   http://localhost:8080/auth/login
   ```

3. You'll be redirected to Microsoft's login page

4. After signing in, you'll be redirected back to the callback URL

5. Check authentication status:
   ```bash
   curl "http://localhost:8080/auth/status?user_id=YOUR_USER_ID"
   ```

### Device Code Flow (for local development)

Device code flow is useful for CLI tools and local development without a browser.

Create a simple test program:

```go
package main

import (
	"context"
	"log"

	"github.com/your-org/opensig/server/internal/auth"
)

func main() {
	config := auth.NewConfigFromEnv()
	tokenStore := auth.NewInMemoryTokenStore()
	
	authService, err := auth.NewService(config, tokenStore)
	if err != nil {
		log.Fatal(err)
	}

	tokenData, err := authService.DeviceCodeFlow(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authenticated: %s (%s)", tokenData.Email, tokenData.UserID)
}
```

Run it and follow the instructions in the console.

## API Endpoints

### GET /auth/login
Initiates OAuth login flow by redirecting to Microsoft sign-in page.

### GET /auth/callback
Handles the OAuth callback from Microsoft. Query parameters:
- `code` - Authorization code
- `state` - CSRF protection token

### GET /auth/status
Check authentication status for a user.

Query parameters:
- `user_id` - User ID to check

Response:
```json
{
  "authenticated": true,
  "user": {
    "id": "user-id",
    "email": "user@example.com"
  },
  "expires_at": "2025-11-17T12:00:00Z"
}
```

### POST /auth/logout
Remove authentication token for a user.

Request body:
```json
{
  "user_id": "user-id"
}
```

## Security Considerations

### Development vs Production

**Development (Current Implementation)**:
- Tokens stored in memory (InMemoryTokenStore)
- Lost on server restart
- Not suitable for production

**Production (TODO)**:
- Use encrypted database storage
- Consider Redis with encryption at rest
- Implement token refresh logic
- Use secure session cookies
- Enable HTTPS only
- Rotate client secrets regularly

### Client Secret Protection

- **Never commit secrets to version control**
- Use environment variables or secret managers (Azure Key Vault, AWS Secrets Manager)
- Rotate secrets regularly
- Use different secrets for dev/staging/production

### Redirect URI Security

- Use HTTPS in production
- Only register necessary redirect URIs
- Validate state parameter to prevent CSRF attacks

## Troubleshooting

### "AZURE_CLIENT_ID is required" error
- Make sure you've set the `AZURE_CLIENT_ID` environment variable
- The value should be the Application (client) ID from Azure Portal

### "Failed to create confidential client" error
- Check that `AZURE_CLIENT_SECRET` is set correctly
- Verify the secret hasn't expired in Azure Portal

### Redirect URI mismatch
- The redirect URI in your environment must exactly match one configured in Azure Portal
- Check for trailing slashes and http vs https

### Token expired
- Tokens typically expire after 1 hour
- Implement token refresh logic using refresh tokens (TODO)

## Next Steps

1. Implement secure token storage for production
2. Add automatic token refresh
3. Integrate with user management system
4. Add support for additional Graph API scopes
5. Implement middleware to validate tokens on protected endpoints

## References

- [Microsoft Identity Platform Documentation](https://docs.microsoft.com/en-us/azure/active-directory/develop/)
- [Microsoft Graph API Documentation](https://docs.microsoft.com/en-us/graph/)
- [OAuth 2.0 Authorization Code Flow](https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-oauth2-auth-code-flow)
- [Device Code Flow](https://docs.microsoft.com/en-us/azure/active-directory/develop/v2-oauth2-device-code)

# OpenSig Windows Signature Agent

Fetches email signature templates from the OpenSig API and writes them to `%APPDATA%\Microsoft\Signatures` for classic Outlook.

## Features

- Authenticates with OpenSig API
- Fetches user-specific signature templates
- Renders templates with user data
- Writes signatures in HTML (.htm), RTF (.rtf), and plain text (.txt) formats
- Logs to stdout and Windows Event Log

## Build

```powershell
cd agent\windows
dotnet build OpenSig.Agent\OpenSig.Agent.csproj -c Release
```

## Run

```powershell
# Set API URL (default: http://localhost:8080)
$env:OPENSIG_API_URL = "https://opensig.example.com"

# Optional: Set user context (for testing without full auth)
$env:OPENSIG_USER_EMAIL = "user@example.com"
$env:OPENSIG_USER_ID = "user-123"
$env:OPENSIG_TENANT_ID = "tenant-123"

# Run the agent
dotnet run --project OpenSig.Agent\OpenSig.Agent.csproj
```

## Configuration

Configure the agent using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENSIG_API_URL` | OpenSig API base URL | `http://localhost:8080` |
| `OPENSIG_USER_EMAIL` | User email (testing only) | - |
| `OPENSIG_USER_ID` | User ID (testing only) | - |
| `OPENSIG_TENANT_ID` | Tenant ID | `default` |

### Server Configuration

The OpenSig API server can be configured with the following environment variable:

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENSIG_SET_DEFAULT_SIGNATURES` | Enable/disable setting default signatures in Outlook | `false` |

When `OPENSIG_SET_DEFAULT_SIGNATURES` is set to `true` or `1`, the agent will:
- Automatically configure the first template as the default signature for new emails and replies/forwards in classic Outlook
- Skip configuration if roaming signatures are detected (to avoid conflicts with Exchange/Microsoft 365 managed signatures)
- Write registry keys to `HKCU\Software\Microsoft\Office\<version>\Outlook\Profiles\Outlook`

## Output

The agent writes signature files to:
- Windows: `%APPDATA%\Microsoft\Signatures`
- Linux/Mac (testing): `~/.config/Microsoft/Signatures`

For each template, it creates:
- `TemplateName.htm` - HTML signature
- `TemplateName.rtf` - RTF signature
- `TemplateName.txt` - Plain text signature

## Logging

The agent logs to:
1. **stdout** - Always enabled with timestamped messages
2. **Windows Event Log** - Logged to Application log under source "OpenSig.Agent" (requires admin privileges to create event source)

## Default Signature Configuration

When the `OPENSIG_SET_DEFAULT_SIGNATURES` feature flag is enabled on the server, the agent will automatically:

1. **Set default signatures** - Configures the first template as the default for new emails and replies/forwards
2. **Detect roaming signatures** - Checks if Microsoft 365/Exchange roaming signatures are enabled and skips configuration to avoid conflicts
3. **Support multiple Outlook versions** - Works with Outlook 2010, 2013, 2016, 2019, and Microsoft 365

**Note:** Setting default signatures requires writing to the Windows registry. The agent will log warnings if:
- Outlook is not installed or hasn't been run yet
- Roaming signatures are enabled
- The user profile hasn't been created

## Next Steps

- Add authentication to fetch user-specific templates from the API using Microsoft Graph
- Handle signature assets (images, logos)
- Add Windows Service support for automatic background sync
- Add update mechanism for signature changes

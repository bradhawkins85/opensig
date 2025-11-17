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

## Next Steps

- Add authentication to fetch user-specific templates from the API using Microsoft Graph
- Set default signature for new/reply messages (classic Outlook registry keys)
- Roaming signatures policy detection (read from registry)
- Handle signature assets (images, logos)
- Add Windows Service support for automatic background sync
- Add update mechanism for signature changes

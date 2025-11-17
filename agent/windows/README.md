# OpenSig Windows Signature Agent (stub)

Writes a sample signature into `%APPDATA%\Microsoft\Signatures` for classic Outlook.

## Build

```powershell
cd agent\windows
dotnet build
dotnet run --project .\OpenSig.Agent\OpenSig.Agent.csproj
```

## Next
- Add authentication to fetch user-specific templates from the API
- Render Liquid/Mustache templates with directory data
- Set default signature for new/reply (classic Outlook)
- Roaming signatures policy detection (read from registry)

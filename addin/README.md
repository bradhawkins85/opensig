# OpenSig Outlook Web Add-in

This directory contains the Outlook Web Add-in for OpenSig, which provides compose-time preview and signature insertion for new Outlook, Outlook Web Access (OWA), and Outlook for Mac.

## Features

### 1. Event-based Compose Activation (LaunchEvent)
The add-in automatically inserts the default signature placeholder `[[signature:default]]` when a new message compose window is opened. This is achieved through the Office.js LaunchEvent API.

**Supported environments:**
- New Outlook (Windows/Mac)
- Outlook Web Access (OWA)
- Outlook for Mac (where LaunchEvents are supported)

**Implementation details:**
- The `OnNewMessageCompose` event triggers the `onNewMessageCompose` function
- The function checks if a signature placeholder already exists to avoid duplicates
- The placeholder is appended at the end of the email body
- Requires Office.js Mailbox API 1.10 or higher

### 2. Manual Signature Insertion
Users can manually insert a signature placeholder using the "Insert signature" button in the ribbon.

### 3. Variant Selection
Users can open the task pane to select from different signature variants (default, marketing, holiday, professional, minimal) and insert the corresponding placeholder.

## File Structure

```
addin/
├── manifest/
│   └── opensig-addin.xml    # Add-in manifest with LaunchEvent configuration
├── src/
│   ├── assets/              # Icons and images
│   ├── functions/
│   │   ├── functions.html   # Function file loader
│   │   └── functions.js     # Function implementations and LaunchEvent handlers
│   └── taskpane/
│       └── taskpane.html    # Task pane UI for variant selection
└── package.json
```

## Development

To run the add-in locally:

```bash
npm start
# or
npm run dev
```

This starts an HTTP server on port 3000 serving the add-in files.

## Manifest Configuration

The manifest includes:
- **Runtimes**: Defines the runtime for LaunchEvent support
- **LaunchEvent ExtensionPoint**: Registers the `OnNewMessageCompose` event
- **MessageComposeCommandSurface**: Ribbon buttons for manual actions
- **Requirements**: Mailbox API 1.10+ for LaunchEvent support

## Testing

To test the add-in:
1. Sideload the manifest in Outlook
2. Open a new compose window - the signature placeholder should be automatically inserted
3. Click the "Insert signature" button to manually insert the default placeholder
4. Click "Choose variant" to open the task pane and select a specific variant

## Signature Placeholders

The add-in inserts placeholders that are later replaced by the OpenSig SMTP relay:
- `[[signature:default]]` - Default signature
- `[[signature:marketing]]` - Marketing signature
- `[[signature:holiday]]` - Holiday signature
- `[[signature:professional]]` - Professional signature
- `[[signature:minimal]]` - Minimal signature

These placeholders are replaced with the actual rendered signatures after the email is sent through the OpenSig relay.

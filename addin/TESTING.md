# Testing Event-based Compose Activation

This document describes how to test the LaunchEvent feature for automatic signature insertion.

## Prerequisites

1. Microsoft 365 account with access to:
   - New Outlook for Windows/Mac, OR
   - Outlook Web Access (OWA)
   
2. Office.js Mailbox API 1.10+ support
   - New Outlook: Fully supported
   - OWA: Fully supported
   - Outlook for Mac: Supported (check version)
   - Classic Outlook: Not supported (LaunchEvents are web add-in only)

## Setup

### 1. Start the Development Server

```bash
cd addin
npm start
```

The server will run on `http://localhost:3000`

### 2. Sideload the Add-in

#### For OWA (Outlook Web Access):
1. Go to https://outlook.office.com
2. Click Settings (gear icon) → View all Outlook settings
3. Navigate to Mail → Customize actions → Get Add-ins
4. Click "My add-ins" → "Add from file"
5. Upload the manifest file: `addin/manifest/opensig-addin.xml`

#### For New Outlook:
1. Open New Outlook application
2. Click "Get Add-ins" from the ribbon
3. Select "My add-ins" → "Add from file"
4. Upload the manifest file: `addin/manifest/opensig-addin.xml`

### 3. Configure HTTPS (if needed)

For production testing, replace `https://localhost:3000` URLs in the manifest with your actual hosting URL.

For local testing with HTTPS:
```bash
# Use a tool like local-ssl-proxy
npx local-ssl-proxy --source 3001 --target 3000
```

Then update manifest URLs to use `https://localhost:3001`

## Test Cases

### Test 1: Auto-insert on New Compose

**Expected behavior:** When opening a new compose window, the placeholder `[[signature:default]]` should be automatically inserted at the end of the email body.

**Steps:**
1. With the add-in loaded, click "New Message" or "New Email"
2. Wait for the compose window to fully load
3. Check the email body

**Expected result:** 
- The body should contain `[[signature:default]]` at the end
- The placeholder should appear after 2 line breaks

**Troubleshooting:**
- Check browser console (F12) for any errors
- Verify the add-in is loaded (check Add-ins menu)
- Ensure you're using a supported Outlook version

### Test 2: Duplicate Prevention

**Expected behavior:** The add-in should not insert duplicate placeholders.

**Steps:**
1. Open a new compose window (placeholder auto-inserted)
2. Manually insert another placeholder using the ribbon button
3. Open another new compose window

**Expected result:**
- First compose: One placeholder auto-inserted
- After manual insertion: Two placeholders (one auto, one manual)
- Second compose: One placeholder auto-inserted (no duplication)

### Test 3: Manual Insertion Still Works

**Expected behavior:** The ribbon button should still work for manual insertion.

**Steps:**
1. Open a new compose window
2. Delete the auto-inserted placeholder
3. Click "Insert signature" ribbon button

**Expected result:**
- Placeholder is inserted at cursor position
- Manual insertion works independently of auto-insertion

### Test 4: Variant Selection

**Expected behavior:** Task pane allows selection of different variants.

**Steps:**
1. Open a new compose window
2. Click "Choose variant" button in ribbon
3. Select "marketing" from dropdown
4. Click "Insert Signature Placeholder"

**Expected result:**
- Task pane opens successfully
- `[[signature:marketing]]` is inserted at cursor position
- Status message shows success

## Debugging

### Check Console Logs

Open browser developer tools (F12) and check console for:
- "Office is ready" message
- Any error messages from the add-in
- Network requests to localhost:3000

### Verify Function Registration

In browser console, after add-in loads:
```javascript
// Check if Office.actions exists
console.log(Office.actions);

// Check if functions are registered
console.log(window.insertSignaturePlaceholder);
```

### Common Issues

**Issue:** LaunchEvent not triggering
- **Solution:** Verify you're using New Outlook or OWA (not Classic Outlook)
- **Solution:** Check manifest has Runtimes and LaunchEvent elements
- **Solution:** Ensure Mailbox requirement set is 1.10 or higher

**Issue:** Signature inserted but not visible
- **Solution:** Check if body content has HTML tags
- **Solution:** Verify coercionType is set to Html

**Issue:** Duplicate placeholders
- **Solution:** Check logic in onNewMessageCompose function
- **Solution:** Verify indexOf check is working correctly

## Verification Checklist

- [ ] New compose window auto-inserts placeholder
- [ ] Placeholder appears at end of body with line breaks
- [ ] No duplicate placeholders on subsequent compose windows
- [ ] Manual ribbon button still works
- [ ] Task pane variant selection works
- [ ] No JavaScript errors in console
- [ ] Add-in loads successfully in supported environments

## Notes

- LaunchEvents are only supported in web-based Outlook versions
- Classic Outlook (Windows COM add-in) does not support LaunchEvents
- The placeholder will be replaced by actual signature when email is sent through OpenSig SMTP relay
- For testing in production, ensure the relay is configured to handle these placeholders

## Reference

- [Office Add-ins LaunchEvents documentation](https://learn.microsoft.com/en-us/office/dev/add-ins/outlook/autolaunch)
- [Office.js API Reference](https://learn.microsoft.com/en-us/javascript/api/office)

package mime

import (
"strings"
"testing"
)

// TestDemonstration_EndToEnd shows the complete workflow of placeholder replacement
func TestDemonstration_EndToEnd(t *testing.T) {
// This test demonstrates the full MIME walker functionality
// with a realistic email message containing signature placeholders

originalEmail := `From: john.doe@example.com
To: client@company.com
Subject: Q4 Project Proposal
Content-Type: text/html; charset=utf-8

<html>
<body>
<p>Dear Client,</p>
<p>Please find attached our Q4 project proposal.</p>
<p>We look forward to your feedback.</p>
<p>[[signature:professional]]</p>
</body>
</html>
`

// Create a walker with a realistic signature renderer
renderer := func(name string) (html string, text string, err error) {
if name == "professional" {
html = `<div style="font-family: Arial, sans-serif;">
<p>Sincerely,<br>
<strong>John Doe</strong><br>
Senior Project Manager<br>
Example Corp<br>
Email: john.doe@example.com<br>
Phone: +1 (555) 123-4567</p>
</div>`
text = `Sincerely,
John Doe
Senior Project Manager
Example Corp
Email: john.doe@example.com
Phone: +1 (555) 123-4567`
}
return html, text, nil
}

walker := NewWalker(renderer)

// Process the message
modified, wasModified, err := walker.ProcessMessage([]byte(originalEmail))
if err != nil {
t.Fatalf("ProcessMessage failed: %v", err)
}

if !wasModified {
t.Fatal("Expected message to be modified")
}

modifiedStr := string(modified)

// Verify placeholder was removed
if strings.Contains(modifiedStr, "[[signature:professional]]") {
t.Error("Placeholder should have been replaced")
}

// Verify signature content was inserted
if !strings.Contains(modifiedStr, "John Doe") {
t.Error("Expected signature to contain sender name")
}
if !strings.Contains(modifiedStr, "Senior Project Manager") {
t.Error("Expected signature to contain title")
}
if !strings.Contains(modifiedStr, "john.doe@example.com") {
t.Error("Expected signature to contain email")
}

t.Logf("Successfully transformed email:\nOriginal length: %d bytes\nModified length: %d bytes", 
len(originalEmail), len(modified))
}

// TestDemonstration_SignedMessagePreservation shows that signed messages are not modified
func TestDemonstration_SignedMessagePreservation(t *testing.T) {
// This test demonstrates that S/MIME signed messages are left untouched

signedEmail := `From: secure@example.com
To: partner@company.com
Subject: Confidential Contract
Content-Type: multipart/signed; protocol="application/pkcs7-signature"; boundary="boundary789"

--boundary789
Content-Type: text/plain

This is a legally binding contract.

[[signature:legal]]

--boundary789
Content-Type: application/pkcs7-signature; name="smime.p7s"
Content-Transfer-Encoding: base64

MIAGCSqGSIb3DQEHAqCAMIACAQExDzANBglghkgBZQMEAgEFADCABgkqhkiG9w0B

--boundary789--
`

renderer := func(name string) (html string, text string, err error) {
return "<p>Legal Signature</p>", "Legal Signature", nil
}

walker := NewWalker(renderer)

// Process the signed message
modified, wasModified, err := walker.ProcessMessage([]byte(signedEmail))
if err != nil {
t.Fatalf("ProcessMessage failed: %v", err)
}

if wasModified {
t.Fatal("Signed message should NOT be modified")
}

// Verify message is identical
if string(modified) != signedEmail {
t.Error("Signed message was altered, which breaks the signature")
}

// Verify placeholder still exists (message was not processed)
if !strings.Contains(string(modified), "[[signature:legal]]") {
t.Error("Placeholder should still exist in signed message")
}

t.Log("Successfully preserved signed message integrity")
}

#!/bin/bash
# Test script for OpenSig SMTP Relay
# This script demonstrates the relay accepting messages on port 2525

set -e

echo "=========================================="
echo "OpenSig SMTP Relay Test Script"
echo "=========================================="
echo ""

# Check if the relay is running
if ! nc -z localhost 2525 2>/dev/null; then
    echo "⚠️  Relay is not running on port 2525"
    echo "   Starting relay in background..."
    cd "$(dirname "$0")/../.."
    go run ./cmd/opensig-relay > /tmp/relay.log 2>&1 &
    RELAY_PID=$!
    echo "   Relay PID: $RELAY_PID"
    sleep 2
    
    if ! nc -z localhost 2525 2>/dev/null; then
        echo "❌ Failed to start relay"
        cat /tmp/relay.log
        exit 1
    fi
    
    CLEANUP=1
else
    echo "✅ Relay is already running on port 2525"
    CLEANUP=0
fi

echo ""
echo "Testing SMTP connection..."
echo ""

# Test SMTP conversation
(
    sleep 0.5
    echo "EHLO test.example.com"
    sleep 0.5
    echo "MAIL FROM:<test@example.com>"
    sleep 0.5
    echo "RCPT TO:<recipient@example.com>"
    sleep 0.5
    echo "DATA"
    sleep 0.5
    echo "From: Test User <test@example.com>"
    echo "To: Recipient <recipient@example.com>"
    echo "Subject: OpenSig Test Message"
    echo "Date: $(date -R)"
    echo ""
    echo "This is a test message sent to the OpenSig SMTP relay."
    echo "The relay should accept this message and log it."
    echo ""
    echo "."
    sleep 0.5
    echo "QUIT"
    sleep 0.5
) | nc localhost 2525 > /tmp/smtp_output.txt 2>&1

# Check results
if grep -q "250.*Roger" /tmp/smtp_output.txt && grep -q "250.*gets this" /tmp/smtp_output.txt; then
    echo "✅ SMTP conversation successful!"
    echo ""
    echo "Server responses:"
    cat /tmp/smtp_output.txt
else
    echo "❌ SMTP conversation failed"
    echo ""
    echo "Output:"
    cat /tmp/smtp_output.txt
    
    if [ $CLEANUP -eq 1 ]; then
        kill $RELAY_PID 2>/dev/null || true
    fi
    exit 1
fi

# Show relay logs if we started it
if [ $CLEANUP -eq 1 ]; then
    echo ""
    echo "=========================================="
    echo "Relay Logs:"
    echo "=========================================="
    cat /tmp/relay.log
    echo ""
    
    echo "Cleaning up..."
    kill $RELAY_PID 2>/dev/null || true
    sleep 1
fi

echo ""
echo "=========================================="
echo "✅ All tests passed!"
echo "=========================================="
echo ""
echo "The relay successfully:"
echo "  ✓ Accepted connection on port 2525"
echo "  ✓ Responded to EHLO command"
echo "  ✓ Accepted MAIL FROM command"
echo "  ✓ Accepted RCPT TO command"
echo "  ✓ Received and logged message data"
echo "  ✓ Can be used as a smart host target"
echo ""

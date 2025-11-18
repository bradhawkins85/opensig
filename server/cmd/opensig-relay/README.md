# OpenSig SMTP Relay

The OpenSig SMTP Relay is a smart host that accepts email messages and can inject email signatures. This document describes how to configure and use the relay.

## Features

- **SMTP Server**: Listens on port 2525 (configurable)
- **STARTTLS Support**: Optional TLS encryption for SMTP connections
- **MTLS Support**: Optional mutual TLS authentication
- **Per-Tenant Authentication**: Validates sender domain against active tenants
- **Comprehensive Logging**: All SMTP operations are logged with session IDs
- **No-op Pass-through**: Currently accepts and logs messages (signature injection coming in future milestone)

## Configuration

The relay is configured through environment variables:

### Basic Configuration

- `SMTP_DOMAIN`: The server's domain name for SMTP greeting (default: `opensig.local`)

### TLS Configuration

To enable STARTTLS/MTLS support, provide certificate files:

- `SMTP_TLS_CERT`: Path to TLS certificate file (PEM format)
- `SMTP_TLS_KEY`: Path to TLS private key file (PEM format)

If these are not set, the relay will run without TLS support.

### Example with TLS

```bash
export SMTP_DOMAIN=mail.example.com
export SMTP_TLS_CERT=/path/to/cert.pem
export SMTP_TLS_KEY=/path/to/key.pem
./opensig-relay
```

### Example without TLS

```bash
export SMTP_DOMAIN=mail.example.com
./opensig-relay
```

## Generating Self-Signed Certificates for Testing

For testing purposes, you can generate a self-signed certificate:

```bash
# Generate private key
openssl genrsa -out relay-key.pem 2048

# Generate self-signed certificate (valid for 365 days)
openssl req -new -x509 -sha256 -key relay-key.pem -out relay-cert.pem -days 365 \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=opensig.local"

# Use the certificates
export SMTP_TLS_CERT=relay-cert.pem
export SMTP_TLS_KEY=relay-key.pem
./opensig-relay
```

## Testing the Relay

### Using telnet

```bash
# Connect to the relay
telnet localhost 2525

# SMTP conversation
EHLO test.example.com
MAIL FROM:<sender@example.com>
RCPT TO:<recipient@example.com>
DATA
Subject: Test Message

This is a test message.
.
QUIT
```

### Using netcat

```bash
(
  echo "EHLO test.example.com"
  echo "MAIL FROM:<sender@example.com>"
  echo "RCPT TO:<recipient@example.com>"
  echo "DATA"
  echo "Subject: Test Message"
  echo ""
  echo "This is a test message."
  echo "."
  echo "QUIT"
) | nc localhost 2525
```

### Using swaks (Swiss Army Knife for SMTP)

```bash
# Install swaks if not available
# apt-get install swaks  # Debian/Ubuntu
# brew install swaks     # macOS

# Send test message
swaks --to recipient@example.com \
      --from sender@example.com \
      --server localhost:2525 \
      --body "Test message body"

# Test with TLS
swaks --to recipient@example.com \
      --from sender@example.com \
      --server localhost:2525 \
      --tls
```

## Configuring as Microsoft 365 Smart Host

To configure the relay as a smart host for Microsoft 365:

1. **Set up the relay** on a server with a public IP or accessible via VPN
2. **Configure connector in Exchange Admin Center**:
   - Go to Mail flow > Connectors
   - Create a new connector from Office 365 to Partner organization
   - Set the smart host address to your relay server (e.g., `mail.example.com:2525`)
   - Configure TLS settings if using STARTTLS
   - Set up appropriate scope and routing rules

3. **Firewall rules**: Ensure port 2525 is accessible from Microsoft 365 IPs

## Per-Tenant Authentication

The relay supports per-tenant authentication using the tenant store:

- Sender email addresses are validated against active tenants
- Domain extraction from email (e.g., `user@example.com` → `example.com`)
- Only active tenants can send mail
- Currently, passwords are accepted in no-op mode (TODO: implement credential validation in M4)

## Logging

All SMTP operations are logged with:
- Session ID for correlation
- Connection source (IP address)
- Authentication attempts and results
- MAIL FROM and RCPT TO commands
- Message size and processing duration
- Error conditions

Example log output:

```
2025/11/17 22:00:00 OpenSig Relay (SMTP smart host) - M3: MTLS listener & message ingest
2025/11/17 22:00:00 Starting SMTP relay server on :2525 (domain: opensig.local)
2025/11/17 22:00:00 Max message size: 10485760 bytes, Max recipients: 100
2025/11/17 22:00:05 [abc123...] New anonymous SMTP connection from 127.0.0.1:54321
2025/11/17 22:00:06 [abc123...] MAIL FROM: <sender@example.com> (anonymous)
2025/11/17 22:00:07 [abc123...] RCPT TO: <recipient@example.com> (anonymous)
2025/11/17 22:00:10 [abc123...] Message received (anonymous): from=<sender@example.com> to=[recipient@example.com] size=456 bytes duration=3.2s
2025/11/17 22:00:10 [abc123...] Session closed (anonymous)
```

## Docker Deployment

The relay is included in the docker-compose setup:

```bash
cd deploy
docker-compose up relay
```

To add TLS certificates to the Docker deployment, mount them as volumes:

```yaml
relay:
  build:
    context: ../server
    target: relay
  ports:
    - "2525:2525"
  environment:
    - SMTP_DOMAIN=mail.example.com
    - SMTP_TLS_CERT=/certs/cert.pem
    - SMTP_TLS_KEY=/certs/key.pem
  volumes:
    - ./certs:/certs:ro
```

## Future Enhancements (M3 Full Implementation)

- MIME parsing and walking
- Placeholder replacement (e.g., `[[signature:default]]`)
- Rules engine integration
- Skip signed/encrypted messages (S/MIME detection)
- Horizontal scaling support
- Metrics and monitoring endpoints
- Message queue integration for async processing

## Troubleshooting

### Port Already in Use

```
Server error: SMTP server error: listen tcp :2525: bind: address already in use
```

Check for existing processes:
```bash
lsof -i :2525
# or
netstat -tlnp | grep 2525
```

### TLS Certificate Errors

Verify certificate files are readable and in PEM format:
```bash
openssl x509 -in cert.pem -text -noout
openssl rsa -in key.pem -check
```

### Connection Refused

- Check firewall rules: `sudo ufw status`
- Verify relay is running: `ps aux | grep opensig-relay`
- Check logs for startup errors

## Security Considerations

- **Production use**: Always use TLS certificates from a trusted CA
- **Network isolation**: Consider running the relay in a private network
- **Rate limiting**: Implement rate limiting at the network/firewall level
- **Monitoring**: Set up alerts for unusual traffic patterns
- **Access control**: Restrict which IPs can connect to the relay

# Developer runbook — local dev

## Prereqs
- Docker & Docker Compose
- Node 20, Go 1.22, .NET 8
- gh CLI (optional)

## Steps
1. `docker compose -f deploy/docker-compose.yml up --build`
2. Visit http://localhost:3000 (web) and http://localhost:8080/healthz (API)

### Environment
Copy `deploy/env/.env.example` to `.env` and adjust.

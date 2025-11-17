# OpenSig (working name)

OpenSig is an open‑source, multi‑tenant email‑signature management system for Microsoft 365, with:

- A **web app** (admin + API) for templates, rules, schedules, RBAC and audits.
- A **Windows Signature Agent** that syncs signatures for **local Outlook preview**.
- An **SMTP/connector relay** (smart host) that **replaces placeholders** such as `[[signature:default]]` **after send**.
- An **Outlook Web Add‑in** for compose‑time preview/insert (new Outlook, OWA, Mac).

This repository contains a skeleton of the solution, including Dockerfiles, a Helm chart, an Outlook add‑in **manifest**,
and CI workflows. Pair it with the PRD in `docs/prd.md`.

> **Stack (suggested)**
> - Server/Relay: Go 1.22
> - Web Admin: Next.js (Node 20 + TypeScript)
> - Outlook Add‑in: Office.js
> - Windows Agent: .NET 8 (C#)
> - Data: Postgres, Redis, S3‑compatible (MinIO for local)

## Quickstart (local, docker compose)

```bash
# 1) Clone and enter
unzip opensig-skeleton.zip -d ./
cd opensig-skeleton

# 2) Build and run
docker compose -f deploy/docker-compose.yml up --build

# 3) Open
# API: http://localhost:8080/healthz
# Web: http://localhost:3000
```

## Quickstart (Kubernetes, Helm)

```bash
# prerequisites: kubectl, helm; a cluster with Ingress
helm install opensig ./deploy/k8s/charts/opensig --values ./deploy/k8s/charts/opensig/values.yaml
```

## Repo Layout

```
server/      # Go API + SMTP relay stubs
web/         # Next.js admin UI stub
addin/       # Outlook Web Add-in manifest + minimal functions
agent/windows# .NET 8 Windows Signature Agent stub
deploy/      # docker-compose + Helm chart
docs/        # PRD and runbooks + ADRs
scripts/     # gh CLI helpers for labels/issues
.github/     # CI workflows + templates
```

## Next steps

- Fill in secrets and environment files (see `deploy/env/.env.example`).
- Use the **scripts** to create labels, milestones, and GitHub issues from the PRD plan:
  ```bash
  bash scripts/create_labels.sh
  bash scripts/create_issues.sh  # requires gh CLI auth
  ```
- Implement features according to milestones M0–M5 in `docs/prd.md`.

## License

AGPL‑3.0 for code; see `LICENSE`.

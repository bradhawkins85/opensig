#!/usr/bin/env bash
set -euo pipefail

# Requires: gh CLI authenticated and repository context set.

# Create milestones
milestones=("M0 - Foundations" "M1 - Windows Agent" "M2 - Outlook Add-in" "M3 - Relay (M365 connectors)" "M4 - Rules/Schedules/Approvals" "M5 - Hardening & Docs")
for m in "${milestones[@]}"; do
  gh api --silent repos/{owner}/{repo}/milestones -f title="$m" >/dev/null 2>&1 || true
done

# Helper to get milestone number by title
ms() {
  gh api repos/{owner}/{repo}/milestones --jq ".[] | select(.title==\"$1\") | .number"
}

create_issue () {
  local title="$1"
  local body="$2"
  local labels="$3"
  local milestone="$4"
  gh issue create --title "$title" --body "$body" --label $labels --milestone "$milestone" >/dev/null
  echo "Created: $title"
}

# M0
create_issue "M0: Tenant model & RBAC scaffolding" \
"Implement multi-tenant data model and role-based access control (Org Admin, Signature Admin, Approver, Auditor).\n\nAcceptance:\n- CRUD tenants\n- Basic roles enforced on a sample endpoint" \
"area/server,security,P1" "$(ms 'M0 - Foundations')"

create_issue "M0: Microsoft Graph auth (Entra ID) — bootstrap" \
"OIDC app registration, minimal scopes, device code login for local dev; store tokens securely.\n\nAcceptance:\n- /auth/login redirects\n- token stored server-side (dev only)" \
"area/server,security,P1" "$(ms 'M0 - Foundations')"

create_issue "M0: Template renderer (Liquid/Mustache) stub" \
"Add renderer module to compile HTML/TXT with placeholders and simple conditionals; sanitize HTML." \
"area/server,P1" "$(ms 'M0 - Foundations')"

# M1
create_issue "M1: Windows Agent — fetch & write signatures" \
"Agent authenticates, pulls assigned templates, writes to %APPDATA%\\Microsoft\\Signatures (.htm/.rtf/.txt) with assets." \
"area/agent,P1" "$(ms 'M1 - Windows Agent')"

create_issue "M1: Windows Agent — default signature (classic Outlook)" \
"Set default signatures for new/reply; feature-flag to avoid altering roaming signatures." \
"area/agent,P2" "$(ms 'M1 - Windows Agent')"

# M2
create_issue "M2: Outlook Web Add-in — insert & variant switch" \
"Ribbon button inserts [[signature:default]]; simple UI to choose variant; function file + task pane." \
"area/addin,P1" "$(ms 'M2 - Outlook Add-in')"

create_issue "M2: Add-in — event-based compose activation (optional)" \
"Investigate LaunchEvent support to auto-insert placeholder on compose (new Outlook/OWA)." \
"area/addin,P2" "$(ms 'M2 - Outlook Add-in')"

# M3
create_issue "M3: SMTP relay — MTLS listener & message ingest" \
"Implement SMTP server with STARTTLS/MTLS, per-tenant auth, logging; no-op pass-through initially." \
"area/server,P1" "$(ms 'M3 - Relay (M365 connectors)')"

create_issue "M3: MIME walker & placeholder replacement" \
"Parse HTML/text parts, replace [[signature:*]] blocks with rendered content; skip S/MIME signed/encrypted messages." \
"area/server,P1" "$(ms 'M3 - Relay (M365 connectors)')"

# M4
create_issue "M4: Rules engine & schedules" \
"Conditions: sender/recipient message type; Schedules: ranges + recurrence; priority + exclusivity." \
"area/server,enhancement,P1" "$(ms 'M4 - Rules/Schedules/Approvals')"

create_issue "M4: Approvals workflow & immutable audit log" \
"Draft->Review->Publish; approvals per area; append-only audit events." \
"area/server,security,P1" "$(ms 'M4 - Rules/Schedules/Approvals')"

# M5
create_issue "M5: Helm chart & production docs" \
"Finalize chart values, documented TLS/ingress, runbooks for admin; security review checklist." \
"docs,area/server,P2" "$(ms 'M5 - Hardening & Docs')"

echo "Issues & milestones created (best-effort). If some already exist, script continues."

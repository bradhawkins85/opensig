#!/usr/bin/env bash
set -euo pipefail

# Requires: gh CLI authenticated and run inside the cloned repo.
# This script:
#   - Ensures milestones M0–M5 exist (by title)
#   - Creates a set of PRD-aligned issues and attaches them to the right milestone

# --- Helpers -----------------------------------------------------------------

ensure_milestone() {
  local title="$1"

  # Check if a milestone with this title already exists
  local existing
  existing=$(gh api repos/:owner/:repo/milestones --jq ".[] | select(.title==\"${title}\") | .number" 2>/dev/null || true)

  if [[ -n "$existing" ]]; then
    echo "Milestone exists: $title (#$existing)"
    return
  fi

  echo "Creating milestone: $title"
  gh api repos/:owner/:repo/milestones -f title="$title" >/dev/null
}

# Create an issue with labels (comma-separated list) and a milestone *title*
create_issue () {
  local title="$1"
  local body="$2"
  local labels_csv="$3"
  local milestone_title="$4"

  # Turn "a,b,c" into separate --label flags
  local label_args=()
  IFS=',' read -ra parts <<< "$labels_csv"
  for lbl in "${parts[@]}"; do
    [[ -n "$lbl" ]] && label_args+=(--label "$lbl")
  done

  gh issue create \
    --title "$title" \
    --body "$body" \
    "${label_args[@]}" \
    --milestone "$milestone_title" >/dev/null

  echo "Created issue: $title (milestone: $milestone_title)"
}

# --- Ensure milestones exist --------------------------------------------------

milestones=(
  "M0 - Foundations"
  "M1 - Windows Agent"
  "M2 - Outlook Add-in"
  "M3 - Relay (M365 connectors)"
  "M4 - Rules/Schedules/Approvals"
  "M5 - Hardening & Docs"
)

for m in "${milestones[@]}"; do
  ensure_milestone "$m"
done

# --- Create issues per milestone ---------------------------------------------

# M0
create_issue \
  "M0: Tenant model & RBAC scaffolding" \
  "Implement multi-tenant data model and role-based access control (Org Admin, Signature Admin, Approver, Auditor).\n\nAcceptance:\n- CRUD tenants\n- Basic roles enforced on a sample endpoint." \
  "area/server,security,P1" \
  "M0 - Foundations"

create_issue \
  "M0: Microsoft Graph auth (Entra ID) — bootstrap" \
  "OIDC app registration, minimal scopes, device code login for local dev; store tokens securely.\n\nAcceptance:\n- /auth/login redirects to Microsoft sign-in\n- Token stored server-side (dev only) and used to call Graph." \
  "area/server,security,P1" \
  "M0 - Foundations"

create_issue \
  "M0: Template renderer (Liquid/Mustache) stub" \
  "Add renderer module to compile HTML/TXT with placeholders and simple conditionals; sanitize HTML to avoid injection.\n\nAcceptance:\n- /v1/preview endpoint renders a template with sample data\n- Basic unit tests cover conditionals and escaping." \
  "area/server,P1" \
  "M0 - Foundations"

# M1
create_issue \
  "M1: Windows Agent — fetch & write signatures" \
  "Agent authenticates, pulls assigned templates, and writes them to %APPDATA%\\Microsoft\\Signatures (.htm/.rtf/.txt) with assets.\n\nAcceptance:\n- Agent writes a per-user signature\n- Logs basic telemetry to stdout / Windows Event Log." \
  "area/agent,P1" \
  "M1 - Windows Agent"

create_issue \
  "M1: Windows Agent — default signature (classic Outlook)" \
  "Set default signatures for new/reply in classic Outlook; feature-flag to avoid altering roaming signatures.\n\nAcceptance:\n- Admin can enable/disable 'set default signatures'\n- When enabled, new/reply use OpenSig signature by default." \
  "area/agent,P2" \
  "M1 - Windows Agent"

# M2
create_issue \
  "M2: Outlook Web Add-in — insert & variant switch" \
  "Ribbon button inserts [[signature:default]]; simple UI to choose variant; function file + task pane wired to server preview.\n\nAcceptance:\n- Button appears on MessageComposeCommandSurface\n- Selected variant placeholder is inserted into body." \
  "area/addin,P1" \
  "M2 - Outlook Add-in"

create_issue \
  "M2: Add-in — event-based compose activation (optional)" \
  "Investigate LaunchEvent support to auto-insert placeholder on compose (new Outlook/OWA) and call API for a preview.\n\nAcceptance:\n- Launch event registered (where supported)\n- Placeholder auto-inserted for eligible mailboxes." \
  "area/addin,P2" \
  "M2 - Outlook Add-in"

# M3
create_issue \
  "M3: SMTP relay — MTLS listener & message ingest" \
  "Implement SMTP server with STARTTLS/MTLS, per-tenant auth, logging; no-op pass-through initially.\n\nAcceptance:\n- Relay accepts messages on :2525\n- Can be configured as smart host target in test environment." \
  "area/server,P1" \
  "M3 - Relay (M365 connectors)"

create_issue \
  "M3: MIME walker & placeholder replacement" \
  "Parse HTML/text parts, replace [[signature:*]] blocks with rendered content; skip S/MIME signed/encrypted messages.\n\nAcceptance:\n- Unit tests with EML fixtures verify correct insertion\n- Signed/encrypted mails are left untouched and tagged." \
  "area/server,P1" \
  "M3 - Relay (M365 connectors)"

# M4
create_issue \
  "M4: Rules engine & schedules" \
  "Conditions: sender/recipient/message type; Schedules: ranges + recurrence; priority + exclusivity.\n\nAcceptance:\n- Rules stored in DB and evaluated for test messages\n- Schedule UI can create/edit basic windows." \
  "area/server,enhancement,P1" \
  "M4 - Rules/Schedules/Approvals"

create_issue \
  "M4: Approvals workflow & immutable audit log" \
  "Draft->Review->Publish; approvals per area; append-only audit events.\n\nAcceptance:\n- Template changes go through a review step\n- Audit log exposes who/what/when with diffs." \
  "area/server,security,P1" \
  "M4 - Rules/Schedules/Approvals"

# M5
create_issue \
  "M5: Helm chart & production docs" \
  "Finalize chart values, documented TLS/ingress patterns, and runbooks for admins; security review checklist.\n\nAcceptance:\n- Helm values support typical small/med. clusters\n- Docs cover rollout, rollback, and basic troubleshooting." \
  "docs,area/server,P2" \
  "M5 - Hardening & Docs"

echo "Issues & milestones created (best-effort). If some already exist, gh will reuse them."

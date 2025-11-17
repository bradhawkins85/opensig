# OpenSig — Product Requirements Document (PRD)

**Doc status:** Draft v1.0  
**Doc owner:** <your-org>  
**Audience:** Engineering, Security, Product, DevOps  
**Last updated:** 2025‑11‑17

## 0) Executive summary

OpenSig is an open‑source, multi‑tenant email‑signature management system for Microsoft 365 (optionally Google Workspace), delivered as a **web application** with:

- a **Windows Signature Agent** that syncs signatures into Outlook for **local preview and default selection**,
- an **email relay** (smart host) that **replaces placeholders** with rendered signatures **after send**,
- **RBAC** for IT/Admin/Marketing workflows,
- a **rules engine** for **schedules**, **targeting**, **exceptions**, and **templates** (HTML + plain text).

## 1) What incumbents do (reference)

- **CodeTwo Email Signatures 365**: server‑side and client‑side modes, Outlook add‑in, preview at compose, rich conditions/exceptions and scheduler.
- **Exclaimer Signatures for M365**: server via connectors and client‑side sync; add‑in with recipient‑aware switching; advanced scheduling/campaigns.
- **Microsoft constraints**: new Outlook uses web add‑ins (no COM); roaming signatures exist; email routing via connectors to smart hosts is supported.

## 2) Goals

1) Central, web‑hosted signature management (RBAC)  
2) Local Outlook preview via Windows Agent; compose preview via Outlook Web Add‑in  
3) Mail relay to replace placeholders post‑send (M365 connectors / SMTP)  
4) Flexible templates and schedules (date/time/recurrence, recipient targeting)  
5) Entra ID (Azure AD) integration via Microsoft Graph

## 3) Non‑goals (v1)

- Full Google Workspace parity (vNext)  
- S/MIME re‑signing or decrypting content

## 4) High‑level architecture

Components: **Web App (Admin + API)**, **Directory Sync Worker**, **Template Renderer**, **Relay (Smart Host)**, **Outlook Web Add‑in**, **Windows Signature Agent**, **Storage** (Postgres/Redis/S3).

## 5) Key requirements (selected)

- **Template system**: HTML + text; placeholders (Liquid/Mustache style), conditional blocks; assets with alt‑text; PNG/JPEG preferred.
- **Rules & schedules**: sender/recipient conditions, message type, time windows, weekly recurrence, priority & exclusivity; exceptions for S/MIME, domains, etc.
- **Windows Agent**: writes to `%APPDATA%\Microsoft\Signatures\` (classic Outlook), cooperates with roaming signatures policies; signed updates.
- **Add‑in**: insert/preview, variant switching, event‑based activation for compose.
- **Relay**: MTLS SMTP smart host; only modifies when placeholder token present; MIME‑aware; skip signed/encrypted; horizontally scalable.
- **RBAC & approvals**: Org Admin, Signature Admin, Approver, Manager, Auditor, End User; immutable audit trail and versioning.
- **Observability**: audit logs; relay volume; agent health.

## 6) Milestones (MVP → GA)

- **M0 Foundations**: Tenant model, RBAC, Graph connect, basic template compile/API
- **M1 Windows Agent**: Sync + install signatures; defaults; roaming policy
- **M2 Outlook Add‑in**: event‑based compose preview; variant picker
- **M3 Relay**: MTLS SMTP, MIME walker, rules engine, connectors E2E
- **M4 Rules/Schedules/Approvals**: full rules model, schedule UI, approvals, audit
- **M5 Hardening/Docs**: security review, packaging, Helm charts, migration guide

See the repository README for dev and deploy details.

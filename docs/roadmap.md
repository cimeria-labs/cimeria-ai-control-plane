# Roadmap

This roadmap is organized around turning Cimeria from a working agent-control prototype into a polished AI SDR operating system.

Captured expansion backlog:

- [Cimeria Expansion Ideas - 2026-05-28](ideas/2026-05-28-cimeria-expansion-ideas.md)
- [Cimeria Expansion Modules Design](superpowers/specs/2026-05-28-cimeria-expansion-modules-design.md)

## Stabilization

- Keep login, workspace creation, daemon registration, and task dispatch green before expanding the product surface.
- Add smoke tests for `/auth/send-code`, workspace creation, daemon registration, lead import, and SDR task creation.
- Make VM deployment reproducible from Git instead of relying on partial manual state.
- Keep runtime logs quiet when no work is available; prefer sleep/wake behavior over noisy polling.

## SDR Pipeline Quality

- Stop the pipeline when Qualificador marks a lead as rejected, low-fit, invalid, or unsafe.
- Synchronize `lead.status`, `state_machine_status`, and `last_event` so UI state matches pipeline state.
- Require structured agent outputs with decision, confidence, rationale, next action, and human approval fields.
- Add import validation for placeholder emails, malformed domains, duplicate leads, and partial Google Places records.
- Store the full activity timeline for each lead so the human can inspect why the pipeline made each decision.

## Integrations

- Promote Google Places configuration into the main backend with clear env validation.
- Add Apollo as a no-send lead source: search preview, approved enrichment, import, and source metadata.
- Add Clay as an enrichment waterfall only when Apollo/internal data is incomplete.
- Add Pipedrive as the CRM handoff layer for qualified leads, deals, notes, activities, and stage sync.
- Refresh or replace stale Cimeria/PAT tokens used by sandbox import tooling.
- Add CRM export targets after lead quality gates are reliable.
- Add webhook/email observability for delivery, bounce, complaint, open, and click events.

## Product Expansion

- Add a Cimeria MCP server so approved tools can be used by Cimeria chat and external agents.
- Add an external agent connector with read-only, draft-only, approval-to-write, and approval-to-send modes.
- Add account intelligence chat grounded in workspace data, leads, issues, CRM state, and agent activity.
- Add mobile/web delivery with WhatsApp-first owner briefs, alerts, and approval requests.
- Add ERP/database/spreadsheet connectors for business diagnostics.
- Add a Business Leak Detector for margin, stock, follow-up, team, and operational bottlenecks.
- Add dynamic executive dashboards that adapt to the company's operation over time.

## SOTA Upgrades

- Build an evaluation harness for agent output quality using golden leads and regression scoring.
- Add human-in-the-loop approval before external outreach is sent.
- Give agents durable memory scoped by workspace, lead, and account.
- Add retrieval over company profile, prior emails, issue history, and product positioning.
- Add cost and latency observability per agent, model, lead, and workspace.
- Add policy-aware autonomy levels: draft-only, suggest, approve-to-send, and autopilot.
- Add background schedulers that wake only for queued work, follow-up windows, or external events.

## Public Repo Quality

- Keep the public repo clean of secrets, backups, generated binaries, database dumps, and VM-specific files.
- Maintain clear docs for architecture, security, roadmap, and local development.
- Prefer small, explainable commits and visible verification in pull requests.

# Cimeria Expansion Modules Design

Date: 2026-05-28
Status: captured for future planning

## Purpose

Capture the expansion direction discussed on May 28 so future planning can turn the ideas into separate implementation plans without losing the product intent.

## Strategic Direction

Cimeria should become an AI operating layer with three major lines:

1. Revenue and SDR control plane.
2. AI operator layer that can use Cimeria tools through chat/MCP.
3. Business diagnostic layer for ERP, spreadsheets, databases, operations, inventory, people, and owner decisions.

The common thread is simple: connect business data, find leverage, recommend action, and execute safely with human approval.

## Module Categories

### Revenue Intelligence

Modules:

- Apollo source and enrichment.
- Clay waterfall enrichment.
- Pipedrive CRM handoff.
- Hunter, Qualificador, Copywriter, Closer, Nurture.
- Approval cockpit.
- Deal quality and ICP score.

Value:

- Finds better leads.
- Prevents blind outreach.
- Produces sales material and next actions.
- Creates CRM records only when lead quality is good enough.

### Cimeria Operator / MCP Layer

Modules:

- Cimeria MCP server.
- External agent connector.
- Tool permission modes.
- Chat-driven workflow execution.
- Audit trail for all tool calls.

Value:

- User can ask Cimeria to do work directly.
- External agents can use Cimeria safely.
- MCP is used for flexible operator/research workflows; stable API remains the production execution layer.

### Account Intelligence Chat

Modules:

- Internal AI chat connected to workspace data.
- Retrieval over leads, issues, activity, CRM, messages, agent output, and settings.
- Action suggestions and approval prompts.
- Email, WhatsApp, task, and cronjob handoff.

Value:

- User can ask questions instead of navigating every screen.
- Answers are grounded in account data.
- Cimeria can turn insight into approved action.

### Mobile and Delivery

Modules:

- Responsive web.
- Mobile-first owner views.
- WhatsApp alerts and command replies.
- Email summaries.
- Scheduled daily/weekly briefings.

Value:

- Owner receives decisions where they already work.
- Cimeria becomes an operating assistant, not only a dashboard.

### Business Diagnostics / Cimeria Ops

Modules:

- ERP/database/spreadsheet connector.
- Company profile analyzer.
- Ticket and margin analyzer.
- Operation quality score.
- Employee/activity evaluator.
- Inventory leak detector.
- Cash leak radar.
- Improvement plan generator.
- Dynamic simple dashboards.

Value:

- Finds money leaking from operations.
- Shows where performance, time, stock, and team execution are weak.
- Produces a consulting-grade improvement plan.
- Sends daily actionable insight to the owner.

## Cross-Cutting Building Blocks

### Connector Hub

Connectors should normalize Apollo, Clay, Pipedrive, ERP, spreadsheets, databases, email, WhatsApp, calendars, and web sources into Cimeria's internal model.

### Unified Business Memory

Cimeria should build a workspace-scoped graph of people, companies, leads, deals, tasks, products, stock, orders, conversations, metrics, and recommendations.

### Approval and Autonomy Modes

Every action should declare its autonomy level:

- read-only;
- draft-only;
- suggest;
- approval-to-send;
- approval-to-write;
- autopilot only for explicitly safe actions.

### ROI Ledger

Every module should try to measure value:

- time saved;
- money found;
- stock corrected;
- leads recovered;
- deals influenced;
- manual work removed;
- owner decisions accelerated.

## SOTA Opportunities

### AI Owner Daily Brief

Daily WhatsApp-first report with the top decisions, risks, money leaks, operational bottlenecks, and approvals.

### Business Leak Detector

System that identifies margin leaks, slow follow-up, ignored high-value leads, inventory issues, product performance problems, and team bottlenecks.

### Consulting Plan Generator

Transforms diagnostics into a client-facing consulting plan with evidence, ROI estimate, rollout steps, and next actions.

### Dynamic Executive Dashboard

Dashboard that adapts automatically as Cimeria learns what matters for the company.

### Agent Skill Marketplace

Reusable agent modules for SDR, ERP analysis, inventory auditing, support triage, finance leak detection, CRM closing, and WhatsApp concierge work.

## Recommended Sequencing

1. Stabilize login, workspace, runtime, schema, and deployment.
2. Finish Apollo no-send integration.
3. Add Pipedrive CRM handoff.
4. Build account intelligence chat over workspace data.
5. Add WhatsApp/mobile owner brief.
6. Add ERP/spreadsheet diagnostic importer.
7. Add Business Leak Detector.
8. Add dynamic dashboards.
9. Add Cimeria MCP server.
10. Add external agent connector.

## Open Questions For Future Specs

1. Should Cimeria Ops be a module inside the current app or a separate product surface?
2. Which first ERP/data source should be supported: spreadsheet upload, Postgres/MySQL, Tiny/Omie/Bling, or generic CSV?
3. Should WhatsApp start as one-way daily brief or two-way command interface?
4. Should MCP come before or after internal chat over account data?
5. Which consulting offer should be packaged first: SDR automation, business diagnostics, or owner daily brief?

## Out Of Scope For This Capture

- Product implementation.
- API contracts.
- UI wireframes.
- Pricing.
- Vendor-specific technical setup.
- Secret management details beyond the approval/autonomy principles.

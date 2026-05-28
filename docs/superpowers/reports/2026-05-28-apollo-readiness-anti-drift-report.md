# Apollo Readiness Anti-Drift Report

Date: 2026-05-28
Status: NO-GO
Scope: analysis only; no Apollo implementation; no product code changes; no real lead email sending.

## Executive Summary

- Overall status: NO-GO
- Login fixed: `/auth/send-code` public smoke test passes; `/auth/verify-code` was not completed because the verification code was not available in-session.
- Workspace creation tested: blocked by unavailable authenticated session.
- Runtime/Hermes tested: partially observed; active daemon and task claims exist, but idle polling is noisy.
- Leads/issues tested: unauthenticated boundary passes; authenticated pages and no-send lead flow were blocked by unavailable authenticated session.
- Apollo readiness: blocked by missing API key, missing connector/dry-run path, schema drift, and provider/import inconsistencies.
- Safe to start Apollo implementation: no; fix the blockers below first.

## Repo and GitHub Baseline

| Check | Result | Evidence |
| --- | --- | --- |
| Local branch | PASS | Active local branch is `codex/apollo-readiness-audit`; created from local `main` after spec and plan commits. |
| Working tree | PASS | `git status -sb` showed a clean working tree on `codex/apollo-readiness-audit`. |
| Ahead/behind | WARN | Audit branch has no upstream yet. Compared to `origin/main`, it includes 3 local commits: spec, plan, and report scaffold. Local `main` is ahead of `origin/main` by 2 commits. |
| Remote | PASS | `origin` points to `https://github.com/cimeria-labs/cimeria-ai-control-plane.git`. |
| Recent commits | PASS | Latest commits: `2d574e6 docs: scaffold Apollo readiness reports`, `4830ea1 docs: add Apollo readiness anti-drift plan`, `cd6c5f3 docs: add Apollo readiness anti-drift spec`, `b3c4ff8 refactor: expose Cimeria icon naming`. |
| GitHub CI | PASS | GitHub repo is public, default branch is `main`, latest origin/main CI run `26328754027` completed successfully for `b3c4ff8`. Current audit branch has not been pushed, so it has no branch CI yet. |

## VM Runtime Baseline

| Check | Result | Evidence |
| --- | --- | --- |
| VM project path | PASS | Active backend process cwd is `/home/opc/swota/multica-main/server`. Additional Git worktree found at `/home/opc/swota-work/multica-main`, currently running the prospector sandbox. |
| VM branch/commit | WARN | Active backend repo is `/home/opc/swota/multica-main` on branch `master` at `d15f5af fix: restore @multica package imports in landing-header`, remote `git@github.com:ferako/swota.git`. This differs from the public repo `cimeria-labs/cimeria-ai-control-plane`. |
| VM dirty state | BLOCKED | Active backend repo is heavily dirty with modified/deleted/untracked files across frontend, backend, CLI, migrations, leads/SDR handlers, generated SQL, parent `.env`, backup directory, logs, and `node_modules`. The running binary cannot be certified as Git-clean. |
| Backend runtime method | PASS | Backend is running as a native ARM64 binary: PID `1149694`, command `./bin/server-arm64`, executable `/home/opc/swota/multica-main/server/bin/server-arm64`, listening on `*:8080`. |
| Frontend runtime method | PASS | Frontend is Dockerized: container `multica-frontend-1`, image `multica-frontend:latest`, bound to `127.0.0.1:3000->3000/tcp`. Additional staging frontend container `multica_stage-frontend-1` is bound to port `13000`. |
| Reverse proxy | PASS | `caddy.service` is active. `/etc/caddy/Caddyfile` routes `app.cimeria.online` `/auth/*`, `/api/*`, `/ws`, and `/uploads/*` to `127.0.0.1:8080`, and all other traffic to `127.0.0.1:3000`. |
| Serving process/container | PASS | Public app path is served by Caddy. Backend: native `server-arm64` PID `1149694`. Frontend: Docker container `multica-frontend-1`. Database: Docker container `multica-postgres-1`. Hermes/daemon: `multica daemon start --foreground` PID `441602`. |

## Secret and Env Status

| Name | Status | Source Checked |
| --- | --- | --- |
| DATABASE_URL | present | Active backend process env and VM env files checked. |
| RESEND_API_KEY | present | Active backend process env and VM env files checked. |
| BACKEND_ORIGIN | present | Active backend process env and VM env files checked. |
| FRONTEND_ORIGIN | present | Active backend process env and VM env files checked. |
| JWT/session/auth secret | present | `JWT_SECRET` is present in active backend process env and VM env files; `SESSION_SECRET`, `AUTH_SECRET`, and `MAGIC_CODE_SECRET` were missing. |
| APOLLO_API_KEY | missing | Missing in active backend process env and checked VM env files. This blocks live Apollo API validation, but not local design/readiness analysis. |

## Product Health

| Flow | Result | Evidence |
| --- | --- | --- |
| POST /auth/send-code public | PASS | VM `curl` to `https://app.cimeria.online/auth/send-code` returned HTTP 200 with `{"message":"Verification code sent"}` on 2026-05-28 06:19 UTC. |
| POST /auth/send-code internal | PASS/WARN | VM `curl` to `http://127.0.0.1:8080/auth/send-code` reached the backend and returned HTTP 429 `please wait before requesting another code` immediately after the public request, which confirms routing and rate limiting. Ports `8081`, `8787`, and `9090` were not serving this route. |
| POST /auth/verify-code | BLOCKED | Verification code was sent to `developercimerio@gmail.com`, but the code was not available in this session. JWT/session token was not obtained. |
| Workspace page/reachability | WARN/BLOCKED | Public frontend root `https://app.cimeria.online` returned HTTP 200 and rendered the Cimeria landing/login entry. Authenticated workspace pages were blocked by unavailable verification code. |
| Runtime registration/status | WARN | VM logs show daemon heartbeats and runtime task claims returning HTTP 200, and `multica daemon start --foreground` is active. However, many runtime IDs are polling/claiming, so idle noise remains a warning. |
| Agents list | BLOCKED | Requires authenticated workspace session; blocked by unavailable verification code. |
| Leads page/API | BLOCKED | Unauthenticated `/api/workspaces` returned expected HTTP 401. Authenticated leads API/page validation blocked by unavailable verification code. |
| Issues page/API | BLOCKED | Requires authenticated workspace session; blocked by unavailable verification code. |
| Lead creates Hunter issue | BLOCKED | No no-send test lead was created because authenticated session was unavailable. |
| Runtime idle noise | WARN | Recent app logs show bursts of daemon heartbeats and repeated `tasks/claim` calls across multiple runtime IDs every few seconds. This does not break auth, but it is noisy. |

## Schema and Pipeline Readiness

| Check | Result | Evidence |
| --- | --- | --- |
| Migrations match generated code expectations | BLOCKED | Public migrations create `lead` and `lead_score_rule`, but generated code and handlers reference `lead_source`, `lead_import_batch`, `lead_curator_rule`, and extra `lead` columns such as `budget`, `authority`, `company_size`, `icp_fit`, `import_batch_id`, `curated_at`, and `curated_by`. VM database contains these objects, so deployed DB has schema not reproducible from public migrations. |
| lead table supports current handlers | BLOCKED | VM database `lead` table includes the generated-code columns, but public `054_leads.up.sql` does not. A clean public migration run would not support the current generated `Lead` model and lead queries. |
| lead_source table exists where handlers expect it | BLOCKED | VM database contains `lead_source`, and handlers/routes/queries exist, but public migrations do not create the table. |
| lead_import_batch table exists where handlers expect it | BLOCKED | VM database contains `lead_import_batch`, and handlers/routes/queries exist, but public migrations do not create the table. |
| Pipeline gates rejected/invalid leads | BLOCKED | `onLeadCreated` skips leads already marked `rejected`, but `onTaskCompleted` advances Hunter -> Qualificador -> Copywriter -> Closer -> Nurture based only on completed agent name. It does not parse agent output or stop on Disqualified/Nurture/invalid/unsafe decisions. |
| Agent outputs are structured enough for decisions | WARN | Agent prompts request structured Markdown sections, but there is no enforced JSON schema with decision, confidence, rationale, next action, and human approval fields for programmatic gating. |

## Apollo Readiness

| Check | Result | Evidence |
| --- | --- | --- |
| APOLLO_API_KEY configured in intended env | BLOCKED | `APOLLO_API_KEY` is missing from the active backend process env and checked VM env files. No live Apollo API call was attempted. |
| API access plan/limits understood | WARN | Official Apollo docs confirm People API Search uses `POST https://api.apollo.io/api/v1/mixed_people/api_search`, is for net-new people, does not return emails/phones, and requires a master API key. Enrichment is separate and credit-bearing. Usage/rate-limit inspection is available via `POST https://api.apollo.io/api/v1/usage_stats/api_usage_stats`, but also requires a master API key. Sources: https://docs.apollo.io/reference/people-api-search, https://docs.apollo.io/reference/people-enrichment, https://docs.apollo.io/reference/bulk-people-enrichment, https://docs.apollo.io/reference/view-api-usage-stats. |
| Secret remains server-side | WARN | No Apollo connector exists yet, so there is no current client-side leak. Future implementation must keep `APOLLO_API_KEY` backend-only and never store it in `lead_source.config`, frontend state, demo assets, logs, or repo files. |
| Source config can store non-secret filters | WARN | `lead_source` supports `provider`, `config`, `auto_approve`, and `enrichment_enabled`; `validLeadSourceProviders` includes `apollo`. However, public migrations do not create the required lead source/import tables, and `CreateLeadImportBatch` currently allows only `csv`, `api`, `form`, and `manual`, so an Apollo batch would be coerced to `api`. |
| Dry-run/import approval path exists or is missing | BLOCKED | There are curator rules and bulk approve/reject actions after leads exist, plus source-level `auto_approve`, but no Apollo search preview/dry-run endpoint was found that lets a human approve candidates before lead creation. This is required before Apollo import. |
| Enrichment is separable from search | WARN | Apollo's API separates search from people/bulk enrichment, and Cimeria has an `enrichment_enabled` source flag. The product still lacks an Apollo connector, enrichment job, webhook/idempotency handling for waterfall, and no-send approval gate. |

## Categorized Findings

### Blockers

- Active deployed backend is not Git-clean or aligned with the public repo: VM runs `/home/opc/swota/multica-main` on `ferako/swota` with many dirty/untracked changes.
- Public migrations do not reproduce the schema required by generated code and deployed DB for `lead_source`, `lead_import_batch`, `lead_curator_rule`, and enriched `lead` columns.
- Authenticated workspace, agents, leads, issues, and no-send lead flow could not be validated without the verification code.
- `APOLLO_API_KEY` is missing in the intended runtime env, so live Apollo validation cannot run.
- No Apollo server-side connector or search preview/dry-run path exists before lead creation.
- SDR pipeline advancement is based on completed agent name, not structured agent decisions, so invalid or disqualified leads can keep moving.

### Bugs

- `CreateLeadImportBatch` provider validation omits `apollo`, while `lead_source` accepts `apollo`; Apollo import batches would lose provider identity unless this is fixed.
- `go test ./...` fails on Windows-specific symlink/path/shell/redaction tests even though handler and SDR packages passed.
- Runtime/daemon idle behavior is noisy, with repeated heartbeat and task-claim activity while no user task is being executed.

### Integration Prerequisites

- Pick one source of truth for deploy: public repo branch -> clean build -> VM pull/deploy, or explicitly document the VM-only private source until it is reconciled.
- Restore or recreate migrations for the schema already present in the VM database and required by generated code.
- Add `APOLLO_API_KEY` only to the backend runtime environment when ready; never commit it or store it in source config.
- Implement Apollo server-side usage check, search preview, candidate cache, import approval, dedupe, and import batch tracking before creating leads.
- Keep enrichment separate from search, controlled by `enrichment_enabled` and a human-approved no-send gate.
- Preserve external Apollo IDs and search/enrichment metadata for dedupe, audit, and re-run safety.

### SDR Quality Improvements

- Enforce structured agent outputs with decision, confidence, rationale, disqualification reason, next action, and human approval requirement.
- Add stop gates after Hunter and Qualificador so rejected, invalid, unsafe, or nurture-only leads do not advance blindly.
- Add no-send verification mode that produces copy, qualification, and closing material without external email delivery.
- Add evaluator fixtures for lead quality, hallucinated company data, tone, and compliance-safe outreach.

### Observability Improvements

- Reduce daemon polling noise with sleep-until-work or longer idle backoff.
- Add request IDs to Apollo import/enrichment events, batch logs, and lead curation actions.
- Avoid broad VM journal reads in future audits because system logs can include infrastructure metadata unrelated to the app.
- Add smoke endpoints or scripts for login, workspace creation, runtime registration, lead creation, Hunter issue creation, and no-send flow.

### SOTA Opportunities

- Lead Intelligence Packet: one normalized profile per lead with source evidence, enrichment confidence, ICP score, objections, and next-best action.
- Waterfall enrichment orchestration: Apollo first, then optional Clay or other enrichers only when data is missing or confidence is low.
- Approval-first SDR cockpit: human reviews candidate leads, generated copy, and send readiness before any external outreach.
- Cost and quota governor for Apollo credits, enrichment calls, LLM tokens, and daemon work.
- Eval dashboard for SDR agents: conversion proxy, lead quality score, hallucination checks, compliance flags, and material quality rubric.

## Final Decision

Decision: NO-GO

NO-GO: code/config/reproducibility fixes are required before Apollo integration and full SDR/Hermes verification can safely resume.

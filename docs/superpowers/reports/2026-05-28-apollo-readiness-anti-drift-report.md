# Apollo Readiness Anti-Drift Report

Date: 2026-05-28
Status: in progress
Scope: analysis only; no Apollo implementation; no product code changes; no real lead email sending.

## Executive Summary

- Overall status: in progress
- Login fixed: not tested
- Workspace creation tested: not tested
- Runtime/Hermes tested: not tested
- Leads/issues tested: not tested
- Apollo readiness: not tested
- Safe to start Apollo implementation: pending

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
| POST /auth/verify-code | pending | pending |
| Workspace page/reachability | pending | pending |
| Runtime registration/status | pending | pending |
| Agents list | pending | pending |
| Leads page/API | pending | pending |
| Issues page/API | pending | pending |
| Lead creates Hunter issue | pending | pending |
| Runtime idle noise | pending | pending |

## Schema and Pipeline Readiness

| Check | Result | Evidence |
| --- | --- | --- |
| Migrations match generated code expectations | pending | pending |
| lead table supports current handlers | pending | pending |
| lead_source table exists where handlers expect it | pending | pending |
| lead_import_batch table exists where handlers expect it | pending | pending |
| Pipeline gates rejected/invalid leads | pending | pending |
| Agent outputs are structured enough for decisions | pending | pending |

## Apollo Readiness

| Check | Result | Evidence |
| --- | --- | --- |
| APOLLO_API_KEY configured in intended env | pending | pending |
| API access plan/limits understood | pending | pending |
| Secret remains server-side | pending | pending |
| Source config can store non-secret filters | pending | pending |
| Dry-run/import approval path exists or is missing | pending | pending |
| Enrichment is separable from search | pending | pending |

## Categorized Findings

### Blockers

- None recorded yet.

### Bugs

- None recorded yet.

### Integration Prerequisites

- None recorded yet.

### SDR Quality Improvements

- None recorded yet.

### Observability Improvements

- None recorded yet.

### SOTA Opportunities

- None recorded yet.

## Final Decision

Decision: pending

Allowed final values: GO, NO-GO, BLOCKED.

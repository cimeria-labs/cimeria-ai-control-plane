# Apollo Readiness Anti-Drift Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Certify that Cimeria's deployed login/workspace/runtime/leads/issues loop is stable and drift-free before Apollo integration work starts.

**Architecture:** This is an analysis-only operational plan. It creates audit reports, gathers local/GitHub/VM evidence, validates the current deployed product loop, checks schema and SDR pipeline readiness, and ends with a go/no-go decision for a later Apollo implementation plan. Product code, SDR architecture, frontend UX, and outbound email behavior are not modified.

**Tech Stack:** Git, GitHub CLI or GitHub web UI, PowerShell on Windows, VM shell access, Go backend, Next.js frontend, PostgreSQL, Caddy, Docker/systemd/native process discovery, curl, ripgrep.

---

## File Structure

- Create: `C:\Users\borac\Documents\cimeria-ai-control-plane\docs\superpowers\reports\2026-05-28-apollo-readiness-anti-drift-report.md`
  - Final human-readable audit with pass/warn/blocker findings.
- Create: `C:\Users\borac\Documents\cimeria-ai-control-plane\docs\superpowers\reports\2026-05-28-apollo-readiness-command-log.md`
  - Redacted command log with sanitized command results.
- Modify: no product source files.

Access rules:

- Use `C:\Users\borac\Documents\Entrega SOTA\ac.txt` only to find the owner-provided VM access method and project paths.
- Do not paste credentials, tokens, passwords, API keys, JWT/session secrets, or raw connection strings into reports.
- For env vars, report only `present`, `missing`, or `empty`.
- If access data is unclear, stop and ask the owner.

### Task 1: Create Report Scaffold

**Files:**
- Create: `C:\Users\borac\Documents\cimeria-ai-control-plane\docs\superpowers\reports\2026-05-28-apollo-readiness-anti-drift-report.md`
- Create: `C:\Users\borac\Documents\cimeria-ai-control-plane\docs\superpowers\reports\2026-05-28-apollo-readiness-command-log.md`

- [ ] **Step 1: Create the reports directory**

Run:

```powershell
New-Item -ItemType Directory -Force -Path 'C:\Users\borac\Documents\cimeria-ai-control-plane\docs\superpowers\reports' | Out-Null
```

Expected: command exits with code 0.

- [ ] **Step 2: Create the main report with fixed sections**

Create `C:\Users\borac\Documents\cimeria-ai-control-plane\docs\superpowers\reports\2026-05-28-apollo-readiness-anti-drift-report.md` with:

```markdown
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
| Local branch | pending | pending |
| Working tree | pending | pending |
| Ahead/behind | pending | pending |
| Remote | pending | pending |
| Recent commits | pending | pending |
| GitHub CI | pending | pending |

## VM Runtime Baseline

| Check | Result | Evidence |
| --- | --- | --- |
| VM project path | pending | pending |
| VM branch/commit | pending | pending |
| VM dirty state | pending | pending |
| Backend runtime method | pending | pending |
| Frontend runtime method | pending | pending |
| Reverse proxy | pending | pending |
| Serving process/container | pending | pending |

## Secret and Env Status

| Name | Status | Source Checked |
| --- | --- | --- |
| DATABASE_URL | pending | pending |
| RESEND_API_KEY | pending | pending |
| BACKEND_ORIGIN | pending | pending |
| FRONTEND_ORIGIN | pending | pending |
| JWT/session/auth secret | pending | pending |
| APOLLO_API_KEY | pending | pending |

## Product Health

| Flow | Result | Evidence |
| --- | --- | --- |
| POST /auth/send-code public | pending | pending |
| POST /auth/send-code internal | pending | pending |
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
```

Expected: report file exists and contains no secrets.

- [ ] **Step 3: Create the command log**

Create `C:\Users\borac\Documents\cimeria-ai-control-plane\docs\superpowers\reports\2026-05-28-apollo-readiness-command-log.md` with:

```markdown
# Apollo Readiness Command Log

Date: 2026-05-28
Scope: redacted command log for anti-drift and readiness analysis.

Rules:

- Do not paste secrets.
- Do not paste raw `.env` files.
- For env vars, record only present, missing, or empty.

## Entries

| Time | Area | Command | Result | Notes |
| --- | --- | --- | --- | --- |
```

Expected: command log exists and contains no secrets.

- [ ] **Step 4: Commit the report scaffold**

Run:

```powershell
git add docs/superpowers/reports/2026-05-28-apollo-readiness-anti-drift-report.md docs/superpowers/reports/2026-05-28-apollo-readiness-command-log.md
git commit -m "docs: scaffold Apollo readiness reports"
```

Expected: commit succeeds with only the two report files.

### Task 2: Capture Local Repo and GitHub Baseline

**Files:**
- Modify: `C:\Users\borac\Documents\cimeria-ai-control-plane\docs\superpowers\reports\2026-05-28-apollo-readiness-anti-drift-report.md`
- Modify: `C:\Users\borac\Documents\cimeria-ai-control-plane\docs\superpowers\reports\2026-05-28-apollo-readiness-command-log.md`

- [ ] **Step 1: Capture local Git state**

Run:

```powershell
Set-Location 'C:\Users\borac\Documents\cimeria-ai-control-plane'
git status -sb
git remote -v
git branch -vv
git log --oneline --decorate -10
git diff --name-only
git diff --stat
```

Expected: commands complete. Record branch, clean/dirty state, ahead/behind count, remote owner/repo, latest commits, and changed paths only.

- [ ] **Step 2: Capture GitHub state**

Run:

```powershell
gh repo view cimeria-labs/cimeria-ai-control-plane --json nameWithOwner,visibility,defaultBranchRef,url
gh run list --repo cimeria-labs/cimeria-ai-control-plane --limit 5
```

Expected: repo metadata and latest workflow runs are visible. If GitHub CLI is not authenticated for `cimeria-labs`, record `BLOCKED: GitHub CLI not authenticated` and inspect the same data in the GitHub web UI.

- [ ] **Step 3: Update report and command log**

Update `Repo and GitHub Baseline` and add command log rows using:

```markdown
| 2026-05-28 HH:MM BRT | repo | git status/log/diff baseline | PASS/WARN/BLOCKED | sanitized result |
| 2026-05-28 HH:MM BRT | github | gh repo view and gh run list | PASS/WARN/BLOCKED | sanitized result |
```

Expected: report has concrete pass/warn/blocker values and no long noisy diffs.

- [ ] **Step 4: Commit baseline**

Run:

```powershell
git add docs/superpowers/reports/2026-05-28-apollo-readiness-anti-drift-report.md docs/superpowers/reports/2026-05-28-apollo-readiness-command-log.md
git commit -m "docs: record repo baseline for Apollo readiness"
```

Expected: commit succeeds if report content changed.

### Task 3: Capture VM Runtime Baseline

**Files:**
- Modify: report and command log files from Task 2.

- [ ] **Step 1: Confirm access file exists**

Run:

```powershell
Test-Path 'C:\Users\borac\Documents\Entrega SOTA\ac.txt'
```

Expected: `True`. If `False`, stop and ask the owner.

- [ ] **Step 2: Connect to VM**

Use only the owner-provided method in `C:\Users\borac\Documents\Entrega SOTA\ac.txt`. Do not write credentials into the report.

Expected: shell access to the VM, or `BLOCKED` with failure category only.

- [ ] **Step 3: Locate deployed project path**

Run on VM:

```bash
pwd
find "$HOME" /opt /srv /var/www -maxdepth 4 -type d \( -name ".git" -o -name "multica-main" -o -name "cimeria-ai-control-plane" -o -name "swota-agent-orchestration-mvp" \) 2>/dev/null | sed 's#/.git$##' | sort -u
```

Expected: active deployed project path is identified, or `BLOCKED: deployed project path not found`.

- [ ] **Step 4: Capture VM Git state**

Run inside the active deployed project directory on VM:

```bash
git status -sb
git remote -v
git branch -vv
git log --oneline --decorate -10
git diff --name-only
git diff --stat
```

Expected: branch, commit, dirty state, remote, and diff status are captured. If active deployment is not a Git worktree, record `WARN: active deployment is not a Git worktree`.

- [ ] **Step 5: Identify runtime method**

Run on VM:

```bash
docker ps --format 'table {{.ID}}\t{{.Image}}\t{{.Names}}\t{{.Status}}\t{{.Ports}}'
ps aux | grep -Ei 'server|multica|cimeria|next|node|caddy|hermes' | grep -v grep
systemctl list-units --type=service --state=running | grep -Ei 'server|multica|cimeria|next|node|caddy|hermes' || true
ss -tulpn 2>/dev/null | grep -Ei ':80|:443|:3000|:8080|:8081|:8787|:9090' || true
```

Expected: backend method is classified as Docker, systemd, native binary, or manual process. Frontend method and reverse proxy are identified.

- [ ] **Step 6: Update report and command log**

Record sanitized values only: paths, process names, container names, branch, commit ID, and clean/dirty state.

Expected: report identifies what is actually serving `app.cimeria.online`.

- [ ] **Step 7: Commit VM baseline**

Run locally:

```powershell
git add docs/superpowers/reports/2026-05-28-apollo-readiness-anti-drift-report.md docs/superpowers/reports/2026-05-28-apollo-readiness-command-log.md
git commit -m "docs: record VM baseline for Apollo readiness"
```

Expected: commit succeeds if report content changed.

### Task 4: Verify Secret and Env Status Safely

**Files:**
- Modify: report and command log files from Task 2.

- [ ] **Step 1: Identify env sources without printing values**

Run on VM:

```bash
systemctl list-units --type=service --all | grep -Ei 'server|multica|cimeria|next|node|hermes' || true
docker ps --format '{{.Names}}' | grep -Ei 'server|multica|cimeria|backend|api|web|next' || true
find /etc /opt /srv /var/www "$HOME" -maxdepth 5 -type f \( -name ".env" -o -name "*.env" -o -name "*.service" \) 2>/dev/null | grep -Ei 'cimeria|multica|sworta|app|server|backend|frontend|web|env|service' || true
```

Expected: candidate env/service file paths are listed by path only.

- [ ] **Step 2: Check process environment safely**

Run on VM:

```bash
PID="$(pgrep -f 'server|multica|cimeria' | head -n 1)"
if [ -n "$PID" ] && [ -r "/proc/$PID/environ" ]; then
  tr '\0' '\n' < "/proc/$PID/environ" | awk -F= '
    BEGIN {
      keys["DATABASE_URL"]=1; keys["RESEND_API_KEY"]=1; keys["BACKEND_ORIGIN"]=1; keys["FRONTEND_ORIGIN"]=1;
      keys["JWT_SECRET"]=1; keys["SESSION_SECRET"]=1; keys["AUTH_SECRET"]=1; keys["MAGIC_CODE_SECRET"]=1; keys["APOLLO_API_KEY"]=1;
    }
    $1 in keys { if (length($2) == 0) print $1 "=empty"; else print $1 "=present"; seen[$1]=1; }
    END { for (k in keys) if (!(k in seen)) print k "=missing"; }' | sort
else
  echo "process_env=unreadable"
fi
```

Expected: statuses only; no secret values.

- [ ] **Step 3: Check env files by key name only**

Run on VM:

```bash
find /etc /opt /srv /var/www "$HOME" -maxdepth 5 -type f \( -name ".env" -o -name "*.env" -o -name "*.service" \) 2>/dev/null \
  | grep -Ei 'cimeria|multica|sworta|app|server|backend|frontend|web|env|service' \
  | while IFS= read -r FILE; do
      echo "checking $(basename "$FILE")"
      for key in DATABASE_URL RESEND_API_KEY BACKEND_ORIGIN FRONTEND_ORIGIN JWT_SECRET SESSION_SECRET AUTH_SECRET MAGIC_CODE_SECRET APOLLO_API_KEY; do
        if grep -qE "^${key}=" "$FILE"; then
          if grep -qE "^${key}=$" "$FILE"; then echo "$key=empty in $(basename "$FILE")"; else echo "$key=present in $(basename "$FILE")"; fi
        else
          echo "$key=missing in $(basename "$FILE")"
        fi
      done
    done
```

Expected: statuses only; no secret values.

- [ ] **Step 4: Update report and command log**

Update `Secret and Env Status` with `present`, `missing`, or `empty` only.

Expected: no secret values in report or command log.

- [ ] **Step 5: Commit env status**

Run locally:

```powershell
git add docs/superpowers/reports/2026-05-28-apollo-readiness-anti-drift-report.md docs/superpowers/reports/2026-05-28-apollo-readiness-command-log.md
git commit -m "docs: record secret status for Apollo readiness"
```

Expected: commit succeeds if report content changed.

### Task 5: Public and Internal Auth Smoke Test

**Files:**
- Modify: report and command log files from Task 2.

- [ ] **Step 1: Test public send-code**

Run from VM:

```bash
curl -i -sS -X POST 'https://app.cimeria.online/auth/send-code' \
  -H 'Content-Type: application/json' \
  -d '{"email":"developercimerio@gmail.com"}'
```

Expected: HTTP 200 or documented auth success response. If HTTP 500 returns, stop Apollo readiness and inspect backend logs.

- [ ] **Step 2: Test internal send-code across likely backend ports**

Run from VM:

```bash
for url in 'http://127.0.0.1:8080/auth/send-code' 'http://127.0.0.1:8081/auth/send-code' 'http://127.0.0.1:8787/auth/send-code' 'http://127.0.0.1:9090/auth/send-code'; do
  echo "Testing $url"
  curl -i -sS -m 10 -X POST "$url" -H 'Content-Type: application/json' -d '{"email":"developercimerio@gmail.com"}' | sed -n '1,20p'
done
```

Expected: one internal URL matches public behavior, or report explains that the internal port was not reachable.

- [ ] **Step 3: Inspect auth logs without hard-coded service names**

Run on VM:

```bash
docker ps --format '{{.Names}}' | grep -Ei 'server|multica|cimeria|backend|api' | while IFS= read -r NAME; do
  echo "logs for $NAME"
  docker logs --since 20m "$NAME" 2>&1 | grep -Ei 'send-code|/auth/send-code|auth|panic|error|500' | tail -n 80
done
systemctl list-units --type=service --state=running --no-legend | awk '{print $1}' | grep -Ei 'server|multica|cimeria|backend|api' | while IFS= read -r UNIT; do
  echo "journal for $UNIT"
  journalctl -u "$UNIT" --since '20 minutes ago' --no-pager | grep -Ei 'send-code|/auth/send-code|auth|panic|error|500' | tail -n 80
done
find /var/log /opt /srv /var/www "$HOME" -maxdepth 5 -type f 2>/dev/null | grep -Ei 'cimeria|multica|server|backend|app|log' | xargs -r grep -Ei 'send-code|/auth/send-code|auth|panic|error|500' | tail -n 80
```

Expected: no panic and no backend 500 for the smoke request.

- [ ] **Step 4: Update report and command log**

Record public/internal status codes and sanitized log evidence.

Expected: send-code is confirmed healthy or blocker is documented.

- [ ] **Step 5: Commit auth smoke results**

Run locally:

```powershell
git add docs/superpowers/reports/2026-05-28-apollo-readiness-anti-drift-report.md docs/superpowers/reports/2026-05-28-apollo-readiness-command-log.md
git commit -m "docs: record auth smoke test for Apollo readiness"
```

Expected: commit succeeds if report content changed.

### Task 6: Product Loop Smoke Test Without External Outreach

**Files:**
- Modify: report and command log files from Task 2.

- [ ] **Step 1: Verify-code with owner-controlled inbox only**

Run from VM or a trusted shell only when the owner can read the code at `developercimerio@gmail.com`:

```bash
read -r -s -p 'Verification code from owner inbox: ' AUTH_CODE
printf '\n'
curl -i -sS -X POST 'https://app.cimeria.online/auth/verify-code' \
  -H 'Content-Type: application/json' \
  -d "$(printf '{"email":"developercimerio@gmail.com","code":"%s"}' "$AUTH_CODE")"
unset AUTH_CODE
```

Expected: HTTP 200 with session/auth success. Do not record JWT/session tokens. If the code is unavailable, mark verify-code as `BLOCKED: code unavailable`.

- [ ] **Step 2: Verify authenticated pages**

Using the authenticated browser session, open the root app, selected workspace, agents page, runtimes page, leads page, and issues page.

Expected: no visible 500s. Agents page includes Hunter, Qualificador, Copywriter, Closer, and Nurture.

- [ ] **Step 3: Create or import one no-send test lead**

Use only this owner-controlled contact data:

```json
{
  "company_name": "Cimeria Apollo Readiness Test",
  "contact_name": "Readiness Test",
  "email": "developercimerio@gmail.com",
  "source": "manual-readiness",
  "notes": "No-send readiness lead for anti-drift audit."
}
```

Expected: lead appears in the leads surface and creates or routes to the expected Hunter issue. No real lead email is sent.

- [ ] **Step 4: Check runtime idle noise**

Run on VM:

```bash
docker ps --format '{{.Names}}' | grep -Ei 'runtime|hermes|server|backend|cimeria|multica' | while IFS= read -r NAME; do
  echo "logs for $NAME"
  docker logs --since 5m "$NAME" 2>&1 | tail -n 120
done
journalctl --since '5 minutes ago' --no-pager | grep -Ei 'hermes|runtime|task|claim|poll|daemon' | tail -n 120
```

Expected: runtime does not produce excessive idle claim/poll noise. If noisy, record warning unless it blocks validation.

- [ ] **Step 5: Update report and command log**

Update product health rows for verify-code, workspace, runtime, agents, leads, issues, lead-to-Hunter issue, and runtime idle noise.

Expected: each row has PASS/WARN/BLOCKED and evidence.

- [ ] **Step 6: Commit product loop results**

Run locally:

```powershell
git add docs/superpowers/reports/2026-05-28-apollo-readiness-anti-drift-report.md docs/superpowers/reports/2026-05-28-apollo-readiness-command-log.md
git commit -m "docs: record product loop smoke test"
```

Expected: commit succeeds if report content changed.

### Task 7: Schema and Pipeline Readiness Audit

**Files:**
- Modify: report and command log files from Task 2.

- [ ] **Step 1: Search local lead/pipeline references**

Run locally:

```powershell
Set-Location 'C:\Users\borac\Documents\cimeria-ai-control-plane'
rg -n "lead_source|lead_import_batch|lead_curator|LeadSource|LeadImport|state_machine_status|last_event|import_batch|enrichment|Hunter|Qualificador|Copywriter|Closer|Nurture" server docs packages
```

Expected: identify handlers, generated queries, migrations, frontend types, and docs that reference pipeline readiness concepts.

- [ ] **Step 2: Compare migrations and generated code**

Run locally:

```powershell
rg -n "CREATE TABLE.*lead|CREATE TABLE.*lead_source|CREATE TABLE.*lead_import_batch|ALTER TABLE.*lead|state_machine_status|last_event|import_batch_id|enrichment" server/migrations
rg -n "type Lead struct|type LeadSource struct|type LeadImportBatch struct|state_machine_status|last_event|import_batch_id|enrichment" server/pkg/db/generated
rg -n "lead_source|lead_import_batch|CreateLead|UpdateLead|ListLeads|CreateLeadSource|CreateLeadImportBatch" server/pkg/db/queries server/internal/handler server/internal/sdr
```

Expected: determine whether migrations create all tables/columns that generated code and handlers expect.

- [ ] **Step 3: Run backend tests**

Run:

```powershell
Set-Location 'C:\Users\borac\Documents\cimeria-ai-control-plane\server'
go test ./...
```

Expected: tests pass, or failures are documented by package/test name and readiness impact.

- [ ] **Step 4: Run frontend typecheck and build validation**

Run:

```powershell
Set-Location 'C:\Users\borac\Documents\cimeria-ai-control-plane'
pnpm.cmd typecheck
pnpm.cmd exec turbo build --env-mode=loose
```

Expected: typecheck and build pass. If they fail, record package/task and first meaningful error only.

- [ ] **Step 5: Check deployed database schema when available**

Run on VM only if `DATABASE_URL=present` and `psql` is available:

```bash
psql "$DATABASE_URL" -Atc "select table_name from information_schema.tables where table_schema='public' and table_name in ('lead','lead_source','lead_import_batch','lead_curator_rule','lead_score_rule') order by table_name;"
psql "$DATABASE_URL" -Atc "select column_name from information_schema.columns where table_schema='public' and table_name='lead' order by ordinal_position;"
```

Expected: table and column names only. Do not print `DATABASE_URL`.

- [ ] **Step 6: Audit SDR progression gates**

Run locally:

```powershell
Get-Content -Path 'C:\Users\borac\Documents\cimeria-ai-control-plane\server\internal\sdr\engine.go' -TotalCount 360
Get-Content -Path 'C:\Users\borac\Documents\cimeria-ai-control-plane\server\internal\handler\sdr_seed.go' -TotalCount 280
```

Expected: determine whether rejected/invalid/unsafe lead decisions stop progression or whether completion always advances to the next agent.

- [ ] **Step 7: Update report and command log**

Update all `Schema and Pipeline Readiness` rows with PASS/WARN/BLOCKED and direct evidence.

Expected: schema and pipeline risks are explicit before Apollo work.

- [ ] **Step 8: Commit schema/pipeline audit**

Run locally:

```powershell
git add docs/superpowers/reports/2026-05-28-apollo-readiness-anti-drift-report.md docs/superpowers/reports/2026-05-28-apollo-readiness-command-log.md
git commit -m "docs: record schema and pipeline readiness audit"
```

Expected: commit succeeds if report content changed.

### Task 8: Apollo Readiness Audit

**Files:**
- Modify: report and command log files from Task 2.

- [ ] **Step 1: Confirm APOLLO_API_KEY status**

Use the status from Task 4. If `APOLLO_API_KEY=missing` or `APOLLO_API_KEY=empty`, record `BLOCKED for live Apollo validation` but continue design/readiness analysis.

Expected: no key value is printed.

- [ ] **Step 2: Confirm official Apollo behavior**

Review:

```text
https://docs.apollo.io/docs/api-overview
https://docs.apollo.io/reference/people-api-search
https://docs.apollo.io/reference/people-enrichment
https://docs.apollo.io/reference/bulk-people-enrichment
https://docs.apollo.io/reference/organization-search
https://docs.apollo.io/reference/view-api-usage-stats
```

Expected: record current facts for auth method, search vs enrichment credit behavior, usage/rate endpoint, and plan/API access. Include links in the report.

- [ ] **Step 3: Check current Cimeria lead source support**

Run locally:

```powershell
Set-Location 'C:\Users\borac\Documents\cimeria-ai-control-plane'
rg -n '"apollo"|validLeadSourceProviders|CreateLeadSource|ListLeadSources|source_id|provider|metadata|external' server packages docs
```

Expected: determine whether Cimeria already recognizes Apollo and whether non-secret config/external metadata can be stored.

- [ ] **Step 4: Evaluate dry-run and approval path**

Run locally:

```powershell
rg -n "dry-run|dry_run|preview|approve|approval|auto_approve|enrichment_enabled|import" server packages docs
```

Expected: determine whether Apollo import can start as preview/dry-run before lead creation and whether source-level approval flags exist.

- [ ] **Step 5: Update report and command log**

Update all `Apollo Readiness` rows with PASS/WARN/BLOCKED and direct evidence.

Expected: future Apollo implementation risks are known.

- [ ] **Step 6: Commit Apollo readiness audit**

Run locally:

```powershell
git add docs/superpowers/reports/2026-05-28-apollo-readiness-anti-drift-report.md docs/superpowers/reports/2026-05-28-apollo-readiness-command-log.md
git commit -m "docs: record Apollo readiness audit"
```

Expected: commit succeeds if report content changed.

### Task 9: Final Go/No-Go Report

**Files:**
- Modify: report and command log files from Task 2.

- [ ] **Step 1: Classify findings**

Move every warning/blocker into exactly one category:

```markdown
### Blockers

- Direct evidence and required fix.

### Bugs

- Direct evidence and user-visible impact.

### Integration Prerequisites

- Required before Apollo implementation starts.

### SDR Quality Improvements

- Pipeline/agent behavior that affects quality but does not block readiness.

### Observability Improvements

- Logging, metrics, or evidence gaps.

### SOTA Opportunities

- High-leverage improvement after the readiness gate.
```

Expected: every warning/blocker is categorized.

- [ ] **Step 2: Set final decision**

Use:

```text
GO if login/workspace/runtime/leads/issues are stable, no schema blocker exists, Apollo can be implemented safely in no-send mode, and missing optional credentials do not block planning.
NO-GO if code or config fixes are required before Apollo can be implemented safely.
BLOCKED if missing access, missing owner-controlled auth code, missing VM access, missing database access, or missing required secret prevents validation.
```

Expected: `Final Decision` contains exactly one value: GO, NO-GO, or BLOCKED.

- [ ] **Step 3: Update executive summary**

Use:

```markdown
- Overall status: GO/NO-GO/BLOCKED
- Login fixed: yes/no/not fully validated
- Workspace creation tested: yes/no/blocked
- Runtime/Hermes tested: yes/no/blocked
- Leads/issues tested: yes/no/blocked
- Apollo readiness: green/yellow/red/blocked
- Safe to start Apollo implementation: yes/no/blocked
```

Expected: summary matches detailed findings.

- [ ] **Step 4: Final verification**

Run:

```powershell
git diff --check
git status -sb
```

Expected: no whitespace errors and only intended report changes before final commit.

- [ ] **Step 5: Commit final report**

Run:

```powershell
git add docs/superpowers/reports/2026-05-28-apollo-readiness-anti-drift-report.md docs/superpowers/reports/2026-05-28-apollo-readiness-command-log.md
git commit -m "docs: finalize Apollo readiness decision"
```

Expected: commit succeeds if report content changed.

- [ ] **Step 6: Report to owner**

Return:

```text
1. Root readiness decision: GO, NO-GO, or BLOCKED
2. Evidence summary
3. Drift findings
4. Product health findings
5. Schema/pipeline findings
6. Apollo readiness findings
7. Files changed
8. Commands run and result summary
9. Current git status -sb
10. Whether it is safe to start Apollo implementation planning
```

Expected: owner can decide whether Apollo implementation planning starts.

## Self-Review Checklist

- The plan covers anti-drift baseline, product health, schema readiness, Apollo readiness, no-send SDR validation, and go/no-go.
- No task modifies product code.
- No task prints secrets.
- VM access uses owner-provided access and stops if access is unclear.
- Apollo validation is no-send.
- Email sending to real leads remains disabled.
- Reports are committed incrementally.

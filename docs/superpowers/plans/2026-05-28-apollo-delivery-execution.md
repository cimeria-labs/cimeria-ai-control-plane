# Apollo Delivery Execution Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver Apollo integrated, tested, functioning in no-send mode, and ready to push.

**Architecture:** Execute the already-written remediation and Apollo no-send plans in a fixed sequence with hard gates. The work lands as small commits on `codex/apollo-readiness-audit`, keeps Apollo secrets backend-only, validates with mocked Apollo tests locally, and validates live Apollo only when `APOLLO_API_KEY` and login verification access are available.

**Tech Stack:** Go 1.26, pgx/sqlc, chi handlers, Apollo REST API, React, TanStack Query, pnpm, Turbo, GitHub branch workflow.

---

## Delivery Answer

Maximum time to push-ready, assuming no dependency blocker:

- **Local push-ready with mocked Apollo tests:** 6 to 8 focused hours.
- **Live Apollo validated on VM:** 8 to 10 focused hours total.
- **If `APOLLO_API_KEY` or login verification code is unavailable:** code can still be push-ready with mocked tests, but live "functioning against Apollo" is blocked.

Hard rule:

- Do not claim live Apollo is functioning until one real `search-preview -> enrich -> import-approved` run succeeds with `no_send=true`.

## Preconditions

Required before starting execution:

- Current branch: `codex/apollo-readiness-audit`.
- Working tree clean.
- Owner confirms it is acceptable to implement on this branch.
- `APOLLO_API_KEY` is available only if live validation is required.
- Owner can provide login verification code for `developercimerio@gmail.com` when VM validation starts.

Check:

```powershell
cd C:\Users\borac\Documents\cimeria-ai-control-plane
git status -sb
git branch --show-current
```

Expected:

- Branch is `codex/apollo-readiness-audit`.
- Working tree has no unstaged or untracked product changes.

## Source Plans

Execute these in order:

1. `docs/superpowers/plans/2026-05-28-apollo-readiness-remediation.md`
2. `docs/superpowers/plans/2026-05-28-apollo-no-send-integration.md`

The first fixes schema/provider drift. The second adds Apollo storage, backend client, handlers, frontend UI, runbook, and validation.

## Stages

### Stage 1: Remediation Gate

**Maximum time:** 90 minutes.

**Purpose:** Make the repo reproducible enough that Apollo code has the tables and provider identity it needs.

**Files are defined in:** `docs/superpowers/plans/2026-05-28-apollo-readiness-remediation.md`

- [ ] **Step 1: Execute remediation Task 1**

Create and run the migration contract test from:

```text
docs/superpowers/plans/2026-05-28-apollo-readiness-remediation.md
Task 1: Add Migration Contract Test
```

Expected:

- Initial test fails before migrations are added.
- Commit: `test: lock lead readiness migration contract`.

- [ ] **Step 2: Execute remediation Task 2**

Add migrations `055`, `056`, `057`, and `058` exactly as specified in the remediation plan.

Expected:

- `go test ./internal/schema -run TestLeadReadinessMigrationsContainGeneratedCodeSchema -count=1` passes.
- Commit: `fix: restore lead readiness migrations`.

- [ ] **Step 3: Execute remediation Task 3**

Fix import batch provider support so `apollo` remains `apollo`.

Expected:

- `go test ./internal/handler -run "TestLeadImportBatchProviderSet" -count=1` passes.
- Commit: `fix: keep Apollo import batch provider identity`.

- [ ] **Step 4: Execute remediation Task 4**

Add Apollo readiness runbook.

Expected:

- Secret-safety scan returns no matches.
- Commit: `docs: add Apollo readiness runbook`.

- [ ] **Step 5: Run remediation verification**

Run:

```powershell
cd C:\Users\borac\Documents\cimeria-ai-control-plane\server
go test ./internal/schema ./internal/handler -count=1
```

Expected:

- PASS.

### Stage 2: Apollo Storage and Backend Client

**Maximum time:** 2 hours.

**Purpose:** Add preview candidate storage and a backend-only Apollo client with tests.

**Files are defined in:** `docs/superpowers/plans/2026-05-28-apollo-no-send-integration.md`

- [ ] **Step 1: Execute Apollo plan Task 1**

Add `lead_import_candidate` migration and sqlc queries.

Expected:

- `go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate` succeeds.
- `go test ./pkg/db/generated ./internal/schema -count=1` passes.
- Commit: `feat: add lead import candidate storage`.

- [ ] **Step 2: Execute Apollo plan Task 2**

Add Apollo backend client and tests.

Expected:

- `go test ./internal/integrations/apollo -count=1` passes.
- Client uses bearer auth.
- Bulk enrichment rejects more than 10 people.
- Phone, personal email, and waterfall flags stay disabled.
- Commit: `feat: add backend Apollo client`.

### Stage 3: Apollo Backend Flow

**Maximum time:** 2.5 hours.

**Purpose:** Add authenticated backend routes for status, search preview, enrichment, and no-send import.

**Files are defined in:** `docs/superpowers/plans/2026-05-28-apollo-no-send-integration.md`

- [ ] **Step 1: Execute Apollo plan Task 3**

Add handler interface, `apollo.go`, tests, and routes.

Expected:

- `GET /api/integrations/apollo/status` returns configured status without exposing secrets.
- `POST /api/integrations/apollo/search-preview` creates candidates but not leads.
- `POST /api/integrations/apollo/enrich` enriches selected candidates.
- `POST /api/integrations/apollo/import-approved` requires `no_send=true`.
- Import creates `source=apollo` leads and no email logs.
- `go test ./internal/handler -run "TestApollo" -count=1` passes.
- Commit: `feat: add Apollo no-send backend flow`.

### Stage 4: Frontend Flow

**Maximum time:** 1.5 hours.

**Purpose:** Make Apollo usable from the Leads page without redesigning the app.

**Files are defined in:** `docs/superpowers/plans/2026-05-28-apollo-no-send-integration.md`

- [ ] **Step 1: Execute Apollo plan Task 4**

Add frontend types, API client methods, queries, and mutations.

Expected:

- `pnpm.cmd typecheck` passes.
- Commit: `feat: expose Apollo no-send API client`.

- [ ] **Step 2: Execute Apollo plan Task 5**

Add Apollo dialog to `packages/views/leads/components/leads-page.tsx`.

Expected:

- User can enter ICP filters, search Apollo preview, select candidates, enrich, and import.
- Apollo button is disabled when backend status says Apollo is not configured.
- `pnpm.cmd typecheck` passes.
- `pnpm.cmd exec turbo build --env-mode=loose` passes.
- Commit: `feat: add Apollo no-send lead import UI`.

### Stage 5: Local Push-Ready Verification

**Maximum time:** 1 hour.

**Purpose:** Prove the branch is push-ready before live VM validation.

- [ ] **Step 1: Run backend focused tests**

Run:

```powershell
cd C:\Users\borac\Documents\cimeria-ai-control-plane\server
go test ./internal/schema ./internal/integrations/apollo ./internal/handler -count=1
```

Expected:

- PASS.

- [ ] **Step 2: Run frontend checks**

Run:

```powershell
cd C:\Users\borac\Documents\cimeria-ai-control-plane
pnpm.cmd typecheck
pnpm.cmd exec turbo build --env-mode=loose
```

Expected:

- PASS.

- [ ] **Step 3: Run full backend suite and classify known failures**

Run:

```powershell
cd C:\Users\borac\Documents\cimeria-ai-control-plane\server
go test ./...
```

Expected:

- PASS, or only previously known Windows-specific failures in daemon symlink/path/shell/redaction tests.
- Any Apollo, handler, schema, generated DB, lead, auth, or workspace failure blocks push.

- [ ] **Step 4: Confirm git is clean**

Run:

```powershell
cd C:\Users\borac\Documents\cimeria-ai-control-plane
git status -sb
git log --oneline --decorate -12
```

Expected:

- Working tree is clean.
- Latest commits are the Apollo delivery commits.

At this point the branch is **ready to push as code**.

### Stage 6: Live Apollo Validation

**Maximum time:** 2 hours after `APOLLO_API_KEY` and login code are available.

**Purpose:** Prove it works against real Apollo and deployed Cimeria without sending email.

- [ ] **Step 1: Confirm Apollo env status without printing values**

Run on VM:

```bash
for key in DATABASE_URL RESEND_API_KEY BACKEND_ORIGIN FRONTEND_ORIGIN JWT_SECRET APOLLO_API_KEY; do
  if [ -z "${!key+x}" ]; then
    printf "%s=missing\n" "$key"
  elif [ -z "${!key}" ]; then
    printf "%s=empty\n" "$key"
  else
    printf "%s=present\n" "$key"
  fi
done
```

Expected:

- All keys required for validation are `present`.

- [ ] **Step 2: Deploy through normal VM path**

Run only after owner approves deploy:

```bash
git fetch origin
git checkout codex/apollo-readiness-audit
git pull --ff-only
cd server
go run ./cmd/migrate up
go build -o bin/server-arm64 ./cmd/server
```

Expected:

- Migrations apply or skip cleanly.
- Backend restarts cleanly using the VM's normal process.
- `/health` returns HTTP 200.

- [ ] **Step 3: Validate login**

Run:

```bash
curl -i -X POST https://app.cimeria.online/auth/send-code \
  -H "Content-Type: application/json" \
  -d '{"email":"developercimerio@gmail.com"}'
```

Expected:

- HTTP 200, or HTTP 429 only if rate-limited after a recent successful request.

After owner provides code:

```bash
TOKEN="$(curl -sS -X POST https://app.cimeria.online/auth/verify-code \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"developercimerio@gmail.com\",\"code\":\"$CODE\"}" | jq -r '.token')"
test -n "$TOKEN" && test "$TOKEN" != "null" && echo "TOKEN=present"
```

Expected:

- `TOKEN=present`.

- [ ] **Step 4: Run requested Apollo test**

Search preview:

```bash
curl -sS -X POST https://app.cimeria.online/api/integrations/apollo/search-preview \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "titles":["Founder","CEO","Head of AI"],
    "person_locations":[],
    "organization_locations":["Brazil"],
    "organization_keywords":["inteligência artificial","artificial intelligence","AI"],
    "seniorities":["founder","owner","c_suite","vp","director","head"],
    "limit":10
  }' > /tmp/apollo-preview.json
jq '{batch_id, candidate_count:(.candidates|length)}' /tmp/apollo-preview.json
```

Expected:

- Candidate count is greater than 0 and at most 10.
- No leads are created by preview.

Enrich approved candidates:

```bash
BATCH_ID="$(jq -r '.batch_id' /tmp/apollo-preview.json)"
CANDIDATE_IDS_JSON="$(jq -c '[.candidates[].id]' /tmp/apollo-preview.json)"
curl -sS -X POST https://app.cimeria.online/api/integrations/apollo/enrich \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"batch_id\":\"$BATCH_ID\",\"candidate_ids\":$CANDIDATE_IDS_JSON}" > /tmp/apollo-enrich.json
jq '{batch_id, enriched:(.candidates|length), emails:[.candidates[].email]}' /tmp/apollo-enrich.json
```

Expected:

- At least one business email appears before import.
- If Apollo returns zero emails, live validation is data-limited and no fake lead is created.

Import no-send:

```bash
ENRICHED_IDS_JSON="$(jq -c '[.candidates[] | select(.email != null) | .id]' /tmp/apollo-enrich.json)"
curl -sS -X POST https://app.cimeria.online/api/integrations/apollo/import-approved \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"batch_id\":\"$BATCH_ID\",\"candidate_ids\":$ENRICHED_IDS_JSON,\"no_send\":true}" > /tmp/apollo-import.json
jq '{imported, skipped, missing_email, duplicates, lead_count:(.leads|length)}' /tmp/apollo-import.json
```

Expected:

- Imported count is greater than 0 when enrichment returned usable emails.
- No outreach email is sent.

- [ ] **Step 5: Prove no-send**

Run:

```bash
psql "$DATABASE_URL" -Atc "
SELECT count(*)
FROM email_log el
JOIN lead l ON l.id = el.lead_id
WHERE l.import_batch_id = '$BATCH_ID';
"
```

Expected:

- `0`.

Run:

```bash
psql "$DATABASE_URL" -Atc "
SELECT l.source, l.status, count(*)
FROM lead l
WHERE l.import_batch_id = '$BATCH_ID'
GROUP BY l.source, l.status
ORDER BY l.source, l.status;
"
```

Expected:

- Imported rows show `source=apollo`.
- Initial status is `captured`, unless curator rules explicitly rejected a lead.

At this point the branch is **ready to push and report as live-validated**.

## Timing Budget

| Stage | Max Time |
| --- | ---: |
| Stage 1 remediation | 1.5h |
| Stage 2 storage/client | 2h |
| Stage 3 backend flow | 2.5h |
| Stage 4 frontend flow | 1.5h |
| Stage 5 local verification | 1h |
| Local push-ready total | 8.5h |
| Stage 6 live VM validation | 2h |
| Live validated total | 10.5h |

If everything is smooth, expected time is closer to 6 to 8 hours. The maximum safe estimate is **one full focused day for push-ready**, plus **up to two hours for live validation** after secrets and login code are available.

## Definition Of Done

Push-ready:

- All remediation tasks committed.
- Apollo storage/client/backend/frontend tasks committed.
- Focused Go tests pass.
- Typecheck passes.
- Turbo build passes.
- Full backend suite has no new Apollo/handler/schema failures.
- Git working tree is clean.

Live functioning:

- `APOLLO_API_KEY=present` in backend runtime env.
- Login works.
- Apollo status returns `configured=true`.
- Search preview returns up to 10 candidates.
- Enrichment returns at least one usable business email.
- Import creates Apollo leads with `no_send=true`.
- `email_log` count for imported leads is `0`.

## Sources

- Apollo People API Search: https://docs.apollo.io/reference/people-api-search
- Apollo People Enrichment: https://docs.apollo.io/reference/people-enrichment
- Apollo Bulk People Enrichment: https://docs.apollo.io/reference/bulk-people-enrichment
- Apollo API usage stats: https://docs.apollo.io/reference/view-api-usage-stats

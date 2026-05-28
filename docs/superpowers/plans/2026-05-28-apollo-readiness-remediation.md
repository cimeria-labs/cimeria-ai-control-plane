# Apollo Readiness Remediation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the public repo reproducible enough to safely start the Apollo no-send integration plan.

**Architecture:** This plan fixes only the first readiness blockers found in the anti-drift audit: missing public migrations, Apollo provider drift in import batches, and missing runbook guardrails. It does not call Apollo, does not add SDR agent behavior, does not deploy to the VM, and does not send email.

**Tech Stack:** Go 1.26, pgx, sqlc-generated query layer, existing SQL migration runner, PowerShell on local Windows, Linux shell on VM for later deployment.

---

## Scope

This is the first remediation slice after `docs/superpowers/reports/2026-05-28-apollo-readiness-anti-drift-report.md`.

Included:

- Restore missing public migrations for lead readiness tables and columns.
- Add tests that keep migrations aligned with generated lead code expectations.
- Fix `CreateLeadImportBatch` provider validation so Apollo batches retain `provider = "apollo"`.
- Add an Apollo readiness runbook with safe, redacted checks.

Excluded:

- Apollo API client.
- Apollo live API validation.
- Clay, Pipedrive, or enrichment waterfall implementation.
- SDR pipeline gating changes.
- VM deployment.
- Any secret creation or secret printing.

## File Structure

- Create: `server/internal/schema/migration_contract_test.go`
  - Guards that the public migration set contains the lead source/import/curator schema and lead enrichment columns expected by generated code.
- Create: `server/migrations/055_lead_enrichment_columns.up.sql`
  - Adds generated-code lead columns that are missing from public `054_leads.up.sql`.
- Create: `server/migrations/055_lead_enrichment_columns.down.sql`
  - Reverses only the columns added by `055`.
- Create: `server/migrations/056_lead_sources.up.sql`
  - Creates `lead_source` with provider, non-secret config, approval, and enrichment flags.
- Create: `server/migrations/056_lead_sources.down.sql`
  - Drops `lead_source`.
- Create: `server/migrations/057_lead_import_batches.up.sql`
  - Creates `lead_import_batch` and adds `lead.import_batch_id`.
- Create: `server/migrations/057_lead_import_batches.down.sql`
  - Drops `lead.import_batch_id` and `lead_import_batch`.
- Create: `server/migrations/058_lead_curator_rules.up.sql`
  - Creates `lead_curator_rule`.
- Create: `server/migrations/058_lead_curator_rules.down.sql`
  - Drops `lead_curator_rule`.
- Create: `server/internal/handler/lead_import_batch_test.go`
  - Guards Apollo provider support in import batch creation logic.
- Modify: `server/internal/handler/lead_import_batch.go`
  - Expands `validBatchProviders` so it is not narrower than lead source provider support.
- Create: `docs/runbooks/apollo-readiness.md`
  - Documents safe local/VM verification and secret handling without exposing values.

## Tasks

### Task 1: Add Migration Contract Test

**Files:**
- Create: `server/internal/schema/migration_contract_test.go`

- [ ] **Step 1: Create the failing migration contract test**

Create `server/internal/schema/migration_contract_test.go`:

```go
package schema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readMigrationSet(t *testing.T) string {
	t.Helper()

	candidates := []string{
		filepath.FromSlash("../../migrations"),
		filepath.FromSlash("server/migrations"),
	}

	var migrationsDir string
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			migrationsDir = candidate
			break
		}
	}
	if migrationsDir == "" {
		t.Fatalf("migrations directory not found from %s", mustGetwd(t))
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("read migrations directory: %v", err)
	}

	var builder strings.Builder
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".up.sql") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			t.Fatalf("read migration %s: %v", name, err)
		}
		builder.WriteString("\n-- ")
		builder.WriteString(name)
		builder.WriteByte('\n')
		builder.Write(content)
		builder.WriteByte('\n')
	}

	return strings.ToLower(builder.String())
}

func mustGetwd(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	return wd
}

func requireSQL(t *testing.T, sql string, required ...string) {
	t.Helper()
	for _, needle := range required {
		if !strings.Contains(sql, strings.ToLower(needle)) {
			t.Fatalf("expected migration SQL to contain %q", needle)
		}
	}
}

func TestLeadReadinessMigrationsContainGeneratedCodeSchema(t *testing.T) {
	sql := readMigrationSet(t)

	requireSQL(t, sql,
		"ADD COLUMN IF NOT EXISTS budget",
		"ADD COLUMN IF NOT EXISTS authority",
		"ADD COLUMN IF NOT EXISTS need",
		"ADD COLUMN IF NOT EXISTS timeline",
		"ADD COLUMN IF NOT EXISTS company_size",
		"ADD COLUMN IF NOT EXISTS industry",
		"ADD COLUMN IF NOT EXISTS pain_points",
		"ADD COLUMN IF NOT EXISTS icp_fit",
		"ADD COLUMN IF NOT EXISTS lead_temperature",
		"ADD COLUMN IF NOT EXISTS curated_at",
		"ADD COLUMN IF NOT EXISTS curated_by",
		"ADD COLUMN IF NOT EXISTS import_batch_id",
		"CREATE TABLE IF NOT EXISTS lead_source",
		"CREATE TABLE IF NOT EXISTS lead_import_batch",
		"CREATE TABLE IF NOT EXISTS lead_curator_rule",
		"provider TEXT NOT NULL",
		"auto_approve BOOLEAN NOT NULL DEFAULT false",
		"enrichment_enabled BOOLEAN NOT NULL DEFAULT true",
		"metadata JSONB NOT NULL DEFAULT '{}'",
	)
}
```

- [ ] **Step 2: Run the test and verify it fails**

Run:

```powershell
cd C:\Users\borac\Documents\cimeria-ai-control-plane\server
go test ./internal/schema -run TestLeadReadinessMigrationsContainGeneratedCodeSchema -count=1
```

Expected: FAIL with a missing SQL substring such as `ADD COLUMN IF NOT EXISTS budget` or `CREATE TABLE IF NOT EXISTS lead_source`.

- [ ] **Step 3: Commit the failing contract test**

Run:

```powershell
git add server\internal\schema\migration_contract_test.go
git commit -m "test: lock lead readiness migration contract"
```

Expected: commit succeeds.

### Task 2: Add Missing Lead Readiness Migrations

**Files:**
- Create: `server/migrations/055_lead_enrichment_columns.up.sql`
- Create: `server/migrations/055_lead_enrichment_columns.down.sql`
- Create: `server/migrations/056_lead_sources.up.sql`
- Create: `server/migrations/056_lead_sources.down.sql`
- Create: `server/migrations/057_lead_import_batches.up.sql`
- Create: `server/migrations/057_lead_import_batches.down.sql`
- Create: `server/migrations/058_lead_curator_rules.up.sql`
- Create: `server/migrations/058_lead_curator_rules.down.sql`

- [ ] **Step 1: Add lead enrichment columns migration**

Create `server/migrations/055_lead_enrichment_columns.up.sql`:

```sql
-- Lead enrichment fields used by generated lead queries and SDR curation.

ALTER TABLE lead
    ADD COLUMN IF NOT EXISTS budget TEXT NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS authority TEXT NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS need TEXT NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS timeline TEXT NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS company_size TEXT NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS industry TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS pain_points TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS icp_fit TEXT NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS lead_temperature TEXT NOT NULL DEFAULT 'cold',
    ADD COLUMN IF NOT EXISTS curated_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS curated_by UUID REFERENCES member(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_lead_workspace_icp_fit ON lead(workspace_id, icp_fit);
CREATE INDEX IF NOT EXISTS idx_lead_workspace_temperature ON lead(workspace_id, lead_temperature);
CREATE INDEX IF NOT EXISTS idx_lead_curated_by ON lead(curated_by);
```

Create `server/migrations/055_lead_enrichment_columns.down.sql`:

```sql
DROP INDEX IF EXISTS idx_lead_curated_by;
DROP INDEX IF EXISTS idx_lead_workspace_temperature;
DROP INDEX IF EXISTS idx_lead_workspace_icp_fit;

ALTER TABLE lead
    DROP COLUMN IF EXISTS curated_by,
    DROP COLUMN IF EXISTS curated_at,
    DROP COLUMN IF EXISTS lead_temperature,
    DROP COLUMN IF EXISTS icp_fit,
    DROP COLUMN IF EXISTS pain_points,
    DROP COLUMN IF EXISTS industry,
    DROP COLUMN IF EXISTS company_size,
    DROP COLUMN IF EXISTS timeline,
    DROP COLUMN IF EXISTS need,
    DROP COLUMN IF EXISTS authority,
    DROP COLUMN IF EXISTS budget;
```

- [ ] **Step 2: Add lead source migration**

Create `server/migrations/056_lead_sources.up.sql`:

```sql
-- Lead sources describe where candidates come from. Config must contain only non-secret filters.

CREATE TABLE IF NOT EXISTS lead_source (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspace(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    provider TEXT NOT NULL CHECK (provider IN (
        'manual',
        'csv',
        'api',
        'form',
        'apollo',
        'hunter',
        'linkedin',
        'referral',
        'website',
        'hubspot',
        'pipedrive'
    )),
    config JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    auto_approve BOOLEAN NOT NULL DEFAULT false,
    enrichment_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_lead_source_workspace ON lead_source(workspace_id);
CREATE INDEX IF NOT EXISTS idx_lead_source_workspace_provider ON lead_source(workspace_id, provider);
```

Create `server/migrations/056_lead_sources.down.sql`:

```sql
DROP INDEX IF EXISTS idx_lead_source_workspace_provider;
DROP INDEX IF EXISTS idx_lead_source_workspace;

DROP TABLE IF EXISTS lead_source;
```

- [ ] **Step 3: Add lead import batch migration**

Create `server/migrations/057_lead_import_batches.up.sql`:

```sql
-- Import batches track preview/import lifecycle and preserve provider metadata.

CREATE TABLE IF NOT EXISTS lead_import_batch (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspace(id) ON DELETE CASCADE,
    source_id UUID REFERENCES lead_source(id) ON DELETE SET NULL,
    file_name TEXT,
    provider TEXT NOT NULL CHECK (provider IN (
        'manual',
        'csv',
        'api',
        'form',
        'apollo',
        'hunter',
        'linkedin',
        'referral',
        'website',
        'hubspot',
        'pipedrive'
    )),
    total_rows INTEGER NOT NULL DEFAULT 0 CHECK (total_rows >= 0),
    imported_count INTEGER NOT NULL DEFAULT 0 CHECK (imported_count >= 0),
    duplicate_count INTEGER NOT NULL DEFAULT 0 CHECK (duplicate_count >= 0),
    rejected_count INTEGER NOT NULL DEFAULT 0 CHECK (rejected_count >= 0),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN (
        'pending',
        'preview',
        'importing',
        'completed',
        'failed',
        'cancelled'
    )),
    error_log TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE lead
    ADD COLUMN IF NOT EXISTS import_batch_id UUID REFERENCES lead_import_batch(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_lead_import_batch_workspace ON lead_import_batch(workspace_id);
CREATE INDEX IF NOT EXISTS idx_lead_import_batch_workspace_status ON lead_import_batch(workspace_id, status);
CREATE INDEX IF NOT EXISTS idx_lead_import_batch_source ON lead_import_batch(source_id);
CREATE INDEX IF NOT EXISTS idx_lead_import_batch_provider ON lead_import_batch(provider);
CREATE INDEX IF NOT EXISTS idx_lead_import_batch_created ON lead_import_batch(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_lead_import_batch_lead ON lead(import_batch_id);
```

Create `server/migrations/057_lead_import_batches.down.sql`:

```sql
DROP INDEX IF EXISTS idx_lead_import_batch_lead;
DROP INDEX IF EXISTS idx_lead_import_batch_created;
DROP INDEX IF EXISTS idx_lead_import_batch_provider;
DROP INDEX IF EXISTS idx_lead_import_batch_source;
DROP INDEX IF EXISTS idx_lead_import_batch_workspace_status;
DROP INDEX IF EXISTS idx_lead_import_batch_workspace;

ALTER TABLE lead DROP COLUMN IF EXISTS import_batch_id;

DROP TABLE IF EXISTS lead_import_batch;
```

- [ ] **Step 4: Add lead curator rule migration**

Create `server/migrations/058_lead_curator_rules.up.sql`:

```sql
-- Curator rules provide deterministic approve/reject/review recommendations.

CREATE TABLE IF NOT EXISTS lead_curator_rule (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspace(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('approve', 'reject', 'review')),
    field TEXT NOT NULL CHECK (field IN (
        'email',
        'company',
        'name',
        'title',
        'industry',
        'company_size',
        'icp_fit',
        'budget',
        'authority',
        'need',
        'timeline'
    )),
    operator TEXT NOT NULL CHECK (operator IN (
        'exists',
        'not_exists',
        'contains',
        'not_contains',
        'eq',
        'ne',
        'gt',
        'gte',
        'lt',
        'lte',
        'regex',
        'domain_in',
        'domain_not_in'
    )),
    value TEXT,
    priority INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    match_count INTEGER NOT NULL DEFAULT 0 CHECK (match_count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_lead_curator_rule_workspace ON lead_curator_rule(workspace_id);
CREATE INDEX IF NOT EXISTS idx_lead_curator_rule_workspace_active ON lead_curator_rule(workspace_id, is_active);
CREATE INDEX IF NOT EXISTS idx_lead_curator_rule_priority ON lead_curator_rule(priority DESC, created_at DESC);
```

Create `server/migrations/058_lead_curator_rules.down.sql`:

```sql
DROP INDEX IF EXISTS idx_lead_curator_rule_priority;
DROP INDEX IF EXISTS idx_lead_curator_rule_workspace_active;
DROP INDEX IF EXISTS idx_lead_curator_rule_workspace;

DROP TABLE IF EXISTS lead_curator_rule;
```

- [ ] **Step 5: Run migration contract test**

Run:

```powershell
cd C:\Users\borac\Documents\cimeria-ai-control-plane\server
go test ./internal/schema -run TestLeadReadinessMigrationsContainGeneratedCodeSchema -count=1
```

Expected: PASS.

- [ ] **Step 6: Run migration command against local/dev database**

Run only when `DATABASE_URL` points to a disposable local/dev database, never production:

```powershell
cd C:\Users\borac\Documents\cimeria-ai-control-plane\server
go run ./cmd/migrate up
```

Expected: `055_lead_enrichment_columns`, `056_lead_sources`, `057_lead_import_batches`, and `058_lead_curator_rules` are applied or skipped if already applied, followed by `Done.`

- [ ] **Step 7: Commit migrations**

Run:

```powershell
git add server\migrations\055_lead_enrichment_columns.up.sql server\migrations\055_lead_enrichment_columns.down.sql server\migrations\056_lead_sources.up.sql server\migrations\056_lead_sources.down.sql server\migrations\057_lead_import_batches.up.sql server\migrations\057_lead_import_batches.down.sql server\migrations\058_lead_curator_rules.up.sql server\migrations\058_lead_curator_rules.down.sql
git commit -m "fix: restore lead readiness migrations"
```

Expected: commit succeeds.

### Task 3: Fix Apollo Import Batch Provider Drift

**Files:**
- Create: `server/internal/handler/lead_import_batch_test.go`
- Modify: `server/internal/handler/lead_import_batch.go`

- [ ] **Step 1: Add failing provider coverage test**

Create `server/internal/handler/lead_import_batch_test.go`:

```go
package handler

import "testing"

func TestLeadImportBatchProviderSetIncludesApollo(t *testing.T) {
	if !validLeadSourceProviders["apollo"] {
		t.Fatal("lead sources must accept apollo provider")
	}
	if !validBatchProviders["apollo"] {
		t.Fatal("lead import batches must accept apollo provider")
	}
}

func TestLeadImportBatchProviderSetCoversLeadSourceProviders(t *testing.T) {
	required := []string{
		"manual",
		"csv",
		"api",
		"form",
		"apollo",
		"hunter",
		"linkedin",
		"referral",
		"website",
		"hubspot",
		"pipedrive",
	}

	for _, provider := range required {
		if !validLeadSourceProviders[provider] {
			t.Fatalf("lead source provider %q is missing", provider)
		}
		if !validBatchProviders[provider] {
			t.Fatalf("lead import batch provider %q is missing", provider)
		}
	}
}
```

- [ ] **Step 2: Run the test and verify it fails**

Run:

```powershell
cd C:\Users\borac\Documents\cimeria-ai-control-plane\server
go test ./internal/handler -run "TestLeadImportBatchProviderSet" -count=1
```

Expected: FAIL with `lead import batches must accept apollo provider` or a missing provider from `validBatchProviders`.

- [ ] **Step 3: Expand import batch provider validation**

In `server/internal/handler/lead_import_batch.go`, replace:

```go
var validBatchProviders = map[string]bool{
    "csv": true, "api": true, "form": true, "manual": true,
}
```

with:

```go
var validBatchProviders = map[string]bool{
	"manual": true,
	"csv": true,
	"api": true,
	"form": true,
	"apollo": true,
	"hunter": true,
	"linkedin": true,
	"referral": true,
	"website": true,
	"hubspot": true,
	"pipedrive": true,
}
```

- [ ] **Step 4: Run the provider test**

Run:

```powershell
cd C:\Users\borac\Documents\cimeria-ai-control-plane\server
go test ./internal/handler -run "TestLeadImportBatchProviderSet" -count=1
```

Expected: PASS.

- [ ] **Step 5: Run handler package tests**

Run:

```powershell
cd C:\Users\borac\Documents\cimeria-ai-control-plane\server
go test ./internal/handler -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit provider fix**

Run:

```powershell
git add server\internal\handler\lead_import_batch.go server\internal\handler\lead_import_batch_test.go
git commit -m "fix: keep Apollo import batch provider identity"
```

Expected: commit succeeds.

### Task 4: Add Apollo Readiness Runbook

**Files:**
- Create: `docs/runbooks/apollo-readiness.md`

- [ ] **Step 1: Create runbook directory and file**

Create `docs/runbooks/apollo-readiness.md`:

```markdown
# Apollo Readiness Runbook

This runbook validates that Cimeria is ready for Apollo no-send integration.

## Safety Rules

- Do not print secrets.
- Do not paste `.env` files into chats, issues, PRs, logs, screenshots, or docs.
- Report secret state only as `present`, `missing`, or `empty`.
- Keep `APOLLO_API_KEY` server-side only.
- Do not store API keys in `lead_source.config`, frontend state, generated demo assets, or repo files.
- Do not send email to real leads during readiness validation.

## Local Repo Checks

Run from the repo root:

```powershell
git status -sb
git log --oneline --decorate -8
git diff --check
```

Expected:

- Working tree is clean or only contains the intended readiness branch changes.
- No whitespace errors from `git diff --check`.

Run backend-focused tests:

```powershell
cd C:\Users\borac\Documents\cimeria-ai-control-plane\server
go test ./internal/schema ./internal/handler -count=1
```

Expected:

- Migration contract tests pass.
- Handler tests pass.

Run frontend/package checks:

```powershell
cd C:\Users\borac\Documents\cimeria-ai-control-plane
pnpm.cmd typecheck
pnpm.cmd exec turbo build --env-mode=loose
```

Expected:

- Typecheck passes.
- Build passes.

## VM Runtime Checks

Run on the VM without printing secret values:

```bash
git status -sb
git remote -v
git branch -vv
git log --oneline --decorate -8
docker ps
ps aux | grep -i "server\|multica\|cimeria" | grep -v grep
```

Expected:

- Active backend process is identified.
- Active frontend process or container is identified.
- Running code source is known.
- Dirty files are either absent or intentionally documented before deploy.

Check required env keys without printing values:

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

Expected before live Apollo validation:

- `DATABASE_URL=present`
- `RESEND_API_KEY=present`
- `BACKEND_ORIGIN=present`
- `FRONTEND_ORIGIN=present`
- `JWT_SECRET=present`
- `APOLLO_API_KEY=present`

## Auth Smoke

Run from the VM:

```bash
curl -i -X POST https://app.cimeria.online/auth/send-code \
  -H "Content-Type: application/json" \
  -d '{"email":"developercimerio@gmail.com"}'
```

Expected:

- HTTP 200 with a verification-code sent response, or HTTP 429 when rate limited after a recent successful request.
- No backend panic.
- No backend HTTP 500 for the request.

## Authenticated No-Send Product Smoke

After the owner provides the verification code, exchange it for an auth token without printing the token:

```bash
TOKEN="$(curl -s -X POST https://app.cimeria.online/auth/verify-code \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"developercimerio@gmail.com\",\"code\":\"$CODE\"}" | jq -r '.token')"
test -n "$TOKEN" && test "$TOKEN" != "null" && echo "TOKEN=present"
```

Expected:

- `TOKEN=present`

Use the token for authenticated checks:

```bash
curl -i https://app.cimeria.online/api/workspaces \
  -H "Authorization: Bearer $TOKEN"
```

Expected:

- HTTP 200 with accessible workspaces, or the expected empty-workspace response if none exists.

## Apollo Readiness Gate

Apollo implementation can start only when all of these are true:

- Public migrations reproduce `lead`, `lead_source`, `lead_import_batch`, and `lead_curator_rule`.
- `CreateLeadImportBatch` preserves `provider = "apollo"`.
- `APOLLO_API_KEY` is present in backend runtime env.
- Login and authenticated workspace smoke tests pass.
- A no-send import path exists or is the explicit next implementation task.
- No real lead email send happens during validation.
```

- [ ] **Step 2: Review runbook for forbidden secret output**

Run:

```powershell
cd C:\Users\borac\Documents\cimeria-ai-control-plane
rg -n "printenv|cat .*\\.env|APOLLO_API_KEY=.*[A-Za-z0-9]" docs\runbooks\apollo-readiness.md
```

Expected: no matches.

- [ ] **Step 3: Commit runbook**

Run:

```powershell
git add docs\runbooks\apollo-readiness.md
git commit -m "docs: add Apollo readiness runbook"
```

Expected: commit succeeds.

### Task 5: Final Verification

**Files:**
- Verify only; no files changed in this task.

- [ ] **Step 1: Run focused backend verification**

Run:

```powershell
cd C:\Users\borac\Documents\cimeria-ai-control-plane\server
go test ./internal/schema ./internal/handler -count=1
```

Expected: PASS.

- [ ] **Step 2: Run public repo build checks**

Run:

```powershell
cd C:\Users\borac\Documents\cimeria-ai-control-plane
pnpm.cmd typecheck
pnpm.cmd exec turbo build --env-mode=loose
```

Expected: both commands PASS.

- [ ] **Step 3: Record known full-Go-suite status**

Run:

```powershell
cd C:\Users\borac\Documents\cimeria-ai-control-plane\server
go test ./...
```

Expected: Either PASS, or the same known Windows-specific failures from `server/internal/daemon/execenv`, `server/internal/daemon/repocache`, and `server/pkg/redact`. If failures differ, stop and investigate before PR.

- [ ] **Step 4: Confirm clean Git state**

Run:

```powershell
cd C:\Users\borac\Documents\cimeria-ai-control-plane
git status -sb
git log --oneline --decorate -8
```

Expected:

- Working tree is clean.
- Latest commits match this plan's task commits.

## Self-Review

Spec coverage:

- Missing public migrations are covered by Tasks 1 and 2.
- Apollo provider/import drift is covered by Task 3.
- Redacted operational checks are covered by Task 4.
- Final local verification is covered by Task 5.

Placeholder scan:

- No task uses placeholder markers, vague future phrases, or unspecified error handling.
- Every file creation task includes exact content.
- Every verification task includes exact commands and expected results.

Type consistency:

- Migration table names match existing generated model names: `lead_source`, `lead_import_batch`, and `lead_curator_rule`.
- Lead column names match generated `db.Lead` fields and `server/pkg/db/queries/lead.sql`.
- Provider values match `validLeadSourceProviders` and the planned `validBatchProviders`.

## Handoff

After this plan is green, write a separate Apollo no-send integration plan. That next plan should add the server-side Apollo client, usage check, people search preview, candidate cache, import approval endpoint, dedupe, and no-send SDR validation.

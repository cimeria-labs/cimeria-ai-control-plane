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
- `APOLLO_API_KEY` reports `present`

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

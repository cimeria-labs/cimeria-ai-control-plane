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
| 2026-05-28 03:14 BRT | repo | `git status -sb`, `git remote -v`, `git branch -vv`, `git log --oneline --decorate -10`, `git diff --name-only`, `git diff --stat` | PASS/WARN | Working tree clean on `codex/apollo-readiness-audit`; branch has no upstream; local audit branch is 3 commits ahead of `origin/main` by ancestry. |
| 2026-05-28 03:14 BRT | github | `gh repo view cimeria-labs/cimeria-ai-control-plane`, `gh run list --repo cimeria-labs/cimeria-ai-control-plane --limit 5` | PASS | Repo is public with default branch `main`; latest origin/main CI run `26328754027` succeeded. |
| 2026-05-28 03:16 BRT | vm | SSH path discovery and VM Git/runtime baseline | PASS/BLOCKED | Connected via owner-provided SSH. Found active backend path `/home/opc/swota/multica-main/server`; active backend repo is dirty and on `ferako/swota`, not public `cimeria-labs/cimeria-ai-control-plane`. |
| 2026-05-28 03:16 BRT | vm-runtime | `ps`, `sudo docker ps`, process cwd checks, Caddyfile inspection, socket scan | PASS | Backend is native `server-arm64` on port `8080`; frontend is Docker container `multica-frontend-1` on `127.0.0.1:3000`; Caddy proxies `app.cimeria.online` to backend/frontend. |
| 2026-05-28 03:18 BRT | env | Backend process env status and VM env file key scan | PASS/BLOCKED | `DATABASE_URL`, `RESEND_API_KEY`, `BACKEND_ORIGIN`, `FRONTEND_ORIGIN`, and `JWT_SECRET` are present; `APOLLO_API_KEY` is missing. Values were not printed. |

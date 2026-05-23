# Security

This repository is safe to publish because it is a clean export without production `.env` files, database dumps, VM logs, generated binaries, or dependency folders.

## Secret Handling

Never commit:

- `.env` files or environment-specific overrides.
- API keys, OAuth secrets, JWT/session secrets, PATs, or runtime tokens.
- Database dumps, production logs, uploaded customer data, or VM backups.
- Generated binaries that were built with environment-specific assumptions.

Use `.env.example` only as a template. Production deployments must provide real values through the host, CI secret store, or deployment platform.

## Required Production Configuration

The backend expects these values to be configured outside Git:

| Variable | Purpose |
| --- | --- |
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | JWT signing secret |
| `BACKEND_ORIGIN` | Public API origin used for links and tracking |
| `FRONTEND_ORIGIN` | Public web origin used for CORS and auth flow |
| `RESEND_API_KEY` | Email delivery for login codes and outreach |
| `RESEND_WEBHOOK_SECRET` | Optional webhook signature validation |
| `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` | Optional Google OAuth |

If `RESEND_API_KEY` is missing, the development path can log verification codes instead of sending email. Production must configure a real email provider.

## History And Rotation

The old working repository should be treated as private/forensic until its history is fully reviewed. Any secret that ever appeared in an old public commit, backup file, VM log, or database dump should be considered compromised and rotated outside this repository.

## Public Repo Audit Checklist

Before publishing a new version:

```bash
git status -sb
rg -l "ghp_|gho_|sk-|RESEND_API_KEY=|JWT_SECRET=|DATABASE_URL=postgres://.*:.*@" .
find . -name ".env*" -o -name "*.bak" -o -name "*.dump" -o -name "*.sqlite" -o -name "*.db"
```

Only `.env.example` should be present.

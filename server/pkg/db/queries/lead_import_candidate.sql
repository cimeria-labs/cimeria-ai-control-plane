-- name: UpsertLeadImportCandidate :one
INSERT INTO lead_import_candidate (
    workspace_id, batch_id, provider, external_id, email, email_status,
    name, company, title, domain, linkedin_url, status, score, payload
) VALUES (
    $1, $2, $3, $4, sqlc.narg('email'), sqlc.narg('email_status'),
    COALESCE(sqlc.narg('name'), ''), COALESCE(sqlc.narg('company'), ''),
    COALESCE(sqlc.narg('title'), ''), COALESCE(sqlc.narg('domain'), ''),
    COALESCE(sqlc.narg('linkedin_url'), ''), COALESCE(sqlc.narg('status'), 'preview'),
    COALESCE(sqlc.narg('score'), 0), COALESCE(sqlc.narg('payload'), '{}'::jsonb)
)
ON CONFLICT (workspace_id, provider, external_id) DO UPDATE SET
    batch_id = EXCLUDED.batch_id,
    email = COALESCE(EXCLUDED.email, lead_import_candidate.email),
    email_status = COALESCE(EXCLUDED.email_status, lead_import_candidate.email_status),
    name = COALESCE(NULLIF(EXCLUDED.name, ''), lead_import_candidate.name),
    company = COALESCE(NULLIF(EXCLUDED.company, ''), lead_import_candidate.company),
    title = COALESCE(NULLIF(EXCLUDED.title, ''), lead_import_candidate.title),
    domain = COALESCE(NULLIF(EXCLUDED.domain, ''), lead_import_candidate.domain),
    linkedin_url = COALESCE(NULLIF(EXCLUDED.linkedin_url, ''), lead_import_candidate.linkedin_url),
    status = EXCLUDED.status,
    score = EXCLUDED.score,
    payload = EXCLUDED.payload,
    updated_at = now()
RETURNING *;

-- name: ListLeadImportCandidates :many
SELECT * FROM lead_import_candidate
WHERE workspace_id = $1
  AND batch_id = $2
ORDER BY score DESC, created_at ASC
LIMIT $3 OFFSET $4;

-- name: GetLeadImportCandidate :one
SELECT * FROM lead_import_candidate
WHERE id = $1 AND workspace_id = $2;

-- name: ListLeadImportCandidatesByIDs :many
SELECT * FROM lead_import_candidate
WHERE workspace_id = $1
  AND id = ANY(sqlc.arg('ids')::uuid[])
ORDER BY created_at ASC;

-- name: UpdateLeadImportCandidateStatus :one
UPDATE lead_import_candidate SET
    status = $3,
    error = sqlc.narg('error'),
    updated_at = now()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: MarkLeadImportCandidateEnriched :one
UPDATE lead_import_candidate SET
    email = sqlc.narg('email'),
    email_status = sqlc.narg('email_status'),
    status = COALESCE(sqlc.narg('status'), 'enriched'),
    enriched_payload = COALESCE(sqlc.narg('enriched_payload'), '{}'::jsonb),
    updated_at = now()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: MarkLeadImportCandidateImported :one
UPDATE lead_import_candidate SET
    status = 'imported',
    lead_id = $3,
    updated_at = now()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

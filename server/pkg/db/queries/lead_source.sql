-- name: CreateLeadSource :one
INSERT INTO lead_source (
    workspace_id, name, slug, provider, config, is_active, auto_approve, enrichment_enabled
) VALUES ($1, $2, $3, $4, COALESCE(sqlc.narg('config'), '{}'::jsonb), COALESCE(sqlc.narg('is_active'), true), COALESCE(sqlc.narg('auto_approve'), false), COALESCE(sqlc.narg('enrichment_enabled'), true))
RETURNING *;

-- name: ListLeadSources :many
SELECT * FROM lead_source
WHERE workspace_id = $1
ORDER BY name;

-- name: GetLeadSource :one
SELECT * FROM lead_source
WHERE id = $1;

-- name: GetLeadSourceBySlug :one
SELECT * FROM lead_source
WHERE workspace_id = $1 AND slug = $2;

-- name: UpdateLeadSource :one
UPDATE lead_source SET
    name = COALESCE(sqlc.narg('name'), name),
    slug = COALESCE(sqlc.narg('slug'), slug),
    provider = COALESCE(sqlc.narg('provider'), provider),
    config = COALESCE(sqlc.narg('config'), config),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    auto_approve = COALESCE(sqlc.narg('auto_approve'), auto_approve),
    enrichment_enabled = COALESCE(sqlc.narg('enrichment_enabled'), enrichment_enabled),
    updated_at = now()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: DeleteLeadSource :exec
DELETE FROM lead_source
WHERE id = $1 AND workspace_id = $2;

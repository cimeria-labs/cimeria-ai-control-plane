-- name: CreateLeadImportBatch :one
INSERT INTO lead_import_batch (
    workspace_id, source_id, file_name, provider, total_rows, status, metadata
) VALUES ($1, sqlc.narg('source_id'), sqlc.narg('file_name'), $2, $3, COALESCE(sqlc.narg('status'), 'pending'), COALESCE(sqlc.narg('metadata'), '{}'::jsonb))
RETURNING *;

-- name: GetLeadImportBatch :one
SELECT * FROM lead_import_batch
WHERE id = $1;

-- name: ListLeadImportBatches :many
SELECT * FROM lead_import_batch
WHERE workspace_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateLeadImportBatch :one
UPDATE lead_import_batch SET
    imported_count = COALESCE(sqlc.narg('imported_count'), imported_count),
    duplicate_count = COALESCE(sqlc.narg('duplicate_count'), duplicate_count),
    rejected_count = COALESCE(sqlc.narg('rejected_count'), rejected_count),
    status = COALESCE(sqlc.narg('status'), status),
    error_log = COALESCE(sqlc.narg('error_log'), error_log),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CreateLeadCuratorRule :one
INSERT INTO lead_curator_rule (
    workspace_id, name, action, field, operator, value, priority, is_active
) VALUES ($1, $2, $3, $4, $5, sqlc.narg('value'), COALESCE(sqlc.narg('priority'), 0), COALESCE(sqlc.narg('is_active'), true))
RETURNING *;

-- name: ListLeadCuratorRules :many
SELECT * FROM lead_curator_rule
WHERE workspace_id = $1
ORDER BY priority DESC, created_at DESC;

-- name: GetLeadCuratorRule :one
SELECT * FROM lead_curator_rule
WHERE id = $1;

-- name: UpdateLeadCuratorRule :one
UPDATE lead_curator_rule SET
    name = COALESCE(sqlc.narg('name'), name),
    action = COALESCE(sqlc.narg('action'), action),
    field = COALESCE(sqlc.narg('field'), field),
    operator = COALESCE(sqlc.narg('operator'), operator),
    value = COALESCE(sqlc.narg('value'), value),
    priority = COALESCE(sqlc.narg('priority'), priority),
    is_active = COALESCE(sqlc.narg('is_active'), is_active),
    match_count = COALESCE(sqlc.narg('match_count'), match_count),
    updated_at = now()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: DeleteLeadCuratorRule :exec
DELETE FROM lead_curator_rule
WHERE id = $1 AND workspace_id = $2;

-- name: IncrementRuleMatchCount :exec
UPDATE lead_curator_rule
SET match_count = match_count + 1, updated_at = now()
WHERE id = $1;

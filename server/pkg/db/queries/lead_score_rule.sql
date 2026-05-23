-- =====================
-- Lead Score Rule CRUD
-- =====================

-- name: ListLeadScoreRules :many
SELECT * FROM lead_score_rule
WHERE workspace_id = $1
ORDER BY event_type;

-- name: GetLeadScoreRule :one
SELECT * FROM lead_score_rule
WHERE workspace_id = $1 AND event_type = $2;

-- name: UpsertLeadScoreRule :one
INSERT INTO lead_score_rule (
    workspace_id, event_type, weight, max_per_email, enabled
) VALUES ($1, $2, $3, sqlc.narg('max_per_email'), sqlc.narg('enabled'))
ON CONFLICT (workspace_id, event_type) DO UPDATE SET
    weight = EXCLUDED.weight,
    max_per_email = COALESCE(EXCLUDED.max_per_email, lead_score_rule.max_per_email),
    enabled = COALESCE(EXCLUDED.enabled, lead_score_rule.enabled),
    updated_at = now()
RETURNING *;

-- name: DeleteLeadScoreRule :exec
DELETE FROM lead_score_rule WHERE id = $1;

-- name: CalculateDynamicScore :one
-- Computes dynamic score for an issue based on email events and workspace rules.
SELECT COALESCE(SUM(least(
    e.cnt, COALESCE(r.max_per_email, 999)
) * r.weight), 0)::int AS dynamic_score
FROM (
    SELECT event_type, COUNT(*) AS cnt
    FROM email_event ev
    JOIN email_log l ON ev.email_log_id = l.id
    WHERE l.issue_id = $1
    GROUP BY event_type
) e
JOIN lead_score_rule r ON e.event_type = r.event_type
WHERE r.workspace_id = $2 AND r.enabled = true;

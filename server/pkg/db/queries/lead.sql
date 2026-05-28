-- name: CreateLead :one
INSERT INTO lead (
    workspace_id, email, name, company, title, source,
    status, score, dynamic_score, assignee_type, assignee_id,
    pipeline_id, state_machine_status, last_event, metadata,
    budget, authority, need, timeline, company_size, industry,
    pain_points, icp_fit, lead_temperature
) VALUES (
    $1, lower($2), COALESCE(sqlc.narg('name'), ''), COALESCE(sqlc.narg('company'), ''),
    COALESCE(sqlc.narg('title'), ''), COALESCE(sqlc.narg('source'), 'manual'),
    COALESCE(sqlc.narg('status'), 'captured'), COALESCE(sqlc.narg('score'), 0),
    COALESCE(sqlc.narg('dynamic_score'), 0), sqlc.narg('assignee_type'),
    sqlc.narg('assignee_id'), sqlc.narg('pipeline_id'),
    COALESCE(sqlc.narg('state_machine_status'), COALESCE(sqlc.narg('status'), 'captured')),
    sqlc.narg('last_event'), COALESCE(sqlc.narg('metadata'), '{}'::jsonb),
    COALESCE(sqlc.narg('budget'), 'unknown'), COALESCE(sqlc.narg('authority'), 'unknown'),
    COALESCE(sqlc.narg('need'), 'unknown'), COALESCE(sqlc.narg('timeline'), 'unknown'),
    COALESCE(sqlc.narg('company_size'), 'unknown'), COALESCE(sqlc.narg('industry'), ''),
    COALESCE(sqlc.narg('pain_points'), ''), COALESCE(sqlc.narg('icp_fit'), 'unknown'),
    COALESCE(sqlc.narg('lead_temperature'), 'cold')
)
ON CONFLICT (workspace_id, email) DO UPDATE SET
    name = COALESCE(NULLIF(EXCLUDED.name, ''), lead.name),
    company = COALESCE(NULLIF(EXCLUDED.company, ''), lead.company),
    title = COALESCE(NULLIF(EXCLUDED.title, ''), lead.title),
    source = EXCLUDED.source,
    updated_at = now()
RETURNING *;

-- name: ListLeads :many
SELECT * FROM lead
WHERE workspace_id = $1
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'))
ORDER BY updated_at DESC
LIMIT $2 OFFSET $3;

-- name: CountLeads :one
SELECT count(*) FROM lead
WHERE workspace_id = $1
  AND (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status'));

-- name: GetLead :one
SELECT * FROM lead
WHERE id = $1;

-- name: GetLeadInWorkspace :one
SELECT * FROM lead
WHERE id = $1 AND workspace_id = $2;

-- name: UpdateLead :one
UPDATE lead SET
    email = COALESCE(lower(sqlc.narg('email')), email),
    name = COALESCE(sqlc.narg('name'), name),
    company = COALESCE(sqlc.narg('company'), company),
    title = COALESCE(sqlc.narg('title'), title),
    source = COALESCE(sqlc.narg('source'), source),
    status = COALESCE(sqlc.narg('status'), status),
    score = COALESCE(sqlc.narg('score'), score),
    dynamic_score = COALESCE(sqlc.narg('dynamic_score'), dynamic_score),
    assignee_type = COALESCE(sqlc.narg('assignee_type'), assignee_type),
    assignee_id = COALESCE(sqlc.narg('assignee_id'), assignee_id),
    pipeline_id = COALESCE(sqlc.narg('pipeline_id'), pipeline_id),
    state_machine_status = COALESCE(sqlc.narg('state_machine_status'), state_machine_status),
    last_event = COALESCE(sqlc.narg('last_event'), last_event),
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    budget = COALESCE(sqlc.narg('budget'), budget),
    authority = COALESCE(sqlc.narg('authority'), authority),
    need = COALESCE(sqlc.narg('need'), need),
    timeline = COALESCE(sqlc.narg('timeline'), timeline),
    company_size = COALESCE(sqlc.narg('company_size'), company_size),
    industry = COALESCE(sqlc.narg('industry'), industry),
    pain_points = COALESCE(sqlc.narg('pain_points'), pain_points),
    icp_fit = COALESCE(sqlc.narg('icp_fit'), icp_fit),
    lead_temperature = COALESCE(sqlc.narg('lead_temperature'), lead_temperature),
    updated_at = now()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: UpdateLeadDynamicScore :one
UPDATE lead SET
    dynamic_score = $3,
    status = CASE WHEN $3 >= 7 AND status NOT IN ('converted', 'cancelled', 'rejected') THEN 'hot' ELSE status END,
    state_machine_status = CASE WHEN $3 >= 7 AND state_machine_status NOT IN ('converted', 'cancelled', 'rejected') THEN 'hot' ELSE state_machine_status END,
    last_event = COALESCE(sqlc.narg('last_event'), last_event),
    updated_at = now()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: DeleteLead :exec
DELETE FROM lead
WHERE id = $1 AND workspace_id = $2;

-- name: CalculateLeadDynamicScore :one
SELECT COALESCE(SUM(least(
    e.cnt, COALESCE(r.max_per_email, 999)
) * r.weight), 0)::int AS dynamic_score
FROM (
    SELECT ev.event_type, COUNT(*) AS cnt
    FROM email_event ev
    JOIN email_log l ON ev.email_log_id = l.id
    WHERE l.lead_id = $1
    GROUP BY ev.event_type
) e
JOIN lead_score_rule r ON e.event_type = r.event_type
WHERE r.workspace_id = $2 AND r.enabled = true;


-- name: SetLeadImportBatch :exec
UPDATE lead SET import_batch_id = $1 WHERE id = $2 AND workspace_id = $3;

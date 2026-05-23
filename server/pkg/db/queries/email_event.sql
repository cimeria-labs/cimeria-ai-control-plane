-- =====================
-- Email Event CRUD
-- =====================

-- name: CreateEmailEvent :one
INSERT INTO email_event (
    email_log_id, event_type, ip_address, user_agent,
    link_url, city, country, metadata
) VALUES (
    $1, $2, sqlc.narg('ip_address'), sqlc.narg('user_agent'),
    sqlc.narg('link_url'), sqlc.narg('city'), sqlc.narg('country'), sqlc.narg('metadata')
) RETURNING *;

-- name: ListEmailEventsByLog :many
SELECT * FROM email_event
WHERE email_log_id = $1
ORDER BY created_at DESC;

-- name: GetLatestEmailEventByLog :one
SELECT * FROM email_event
WHERE email_log_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: GetEmailEventCountsByLog :one
SELECT
    COUNT(*) FILTER (WHERE event_type = 'opened') AS open_count,
    COUNT(*) FILTER (WHERE event_type = 'clicked') AS click_count,
    COUNT(*) FILTER (WHERE event_type = 'replied') AS reply_count,
    COUNT(*) FILTER (WHERE event_type = 'bounced') AS bounce_count,
    COUNT(*) FILTER (WHERE event_type = 'complained') AS complained_count,
    COUNT(*) FILTER (WHERE event_type = 'unsubscribed') AS unsubscribed_count
FROM email_event
WHERE email_log_id = $1;

-- name: GetEmailEventsByIssue :many
-- All events for all emails linked to an issue.
SELECT e.* FROM email_event e
JOIN email_log l ON e.email_log_id = l.id
WHERE l.issue_id = $1
ORDER BY e.created_at DESC;

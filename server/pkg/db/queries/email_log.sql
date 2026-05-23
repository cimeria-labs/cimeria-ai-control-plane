-- =====================
-- Email Log CRUD
-- =====================

-- name: CreateEmailLog :one
INSERT INTO email_log (
    workspace_id, issue_id, lead_id, sender_id, sender_type,
    "to", subject, body_preview, resend_id, email_type,
    status, tracking_enabled
) VALUES (
    $1, sqlc.narg('issue_id'), sqlc.narg('lead_id'), sqlc.narg('sender_id'), sqlc.narg('sender_type'),
    $2, $3, sqlc.narg('body_preview'), sqlc.narg('resend_id'), $4,
    sqlc.narg('status'), sqlc.narg('tracking_enabled')
) RETURNING *;

-- name: GetEmailLog :one
SELECT * FROM email_log WHERE id = $1;

-- name: GetEmailLogByResendID :one
SELECT * FROM email_log WHERE resend_id = $1;

-- name: ListEmailLogsByIssue :many
SELECT * FROM email_log
WHERE issue_id = $1
ORDER BY sent_at DESC;

-- name: ListEmailLogsByLead :many
SELECT * FROM email_log
WHERE lead_id = $1
ORDER BY sent_at DESC;

-- name: ListEmailLogsByWorkspace :many
SELECT * FROM email_log
WHERE workspace_id = $1
ORDER BY sent_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateEmailLogStatus :one
UPDATE email_log SET
    status = COALESCE(sqlc.narg('status'), status),
    resend_id = COALESCE(sqlc.narg('resend_id'), resend_id),
    opened_at = COALESCE(sqlc.narg('opened_at'), opened_at),
    clicked_at = COALESCE(sqlc.narg('clicked_at'), clicked_at),
    replied_at = COALESCE(sqlc.narg('replied_at'), replied_at),
    bounce_reason = COALESCE(sqlc.narg('bounce_reason'), bounce_reason),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetEmailLogStatsByIssue :one
SELECT
    COUNT(*) FILTER (WHERE status = 'sent') AS total_sent,
    COUNT(*) FILTER (WHERE opened_at IS NOT NULL) AS total_opened,
    COUNT(*) FILTER (WHERE clicked_at IS NOT NULL) AS total_clicked,
    COUNT(*) FILTER (WHERE replied_at IS NOT NULL) AS total_replied,
    COUNT(*) FILTER (WHERE status = 'bounced') AS total_bounced
FROM email_log
WHERE issue_id = $1;

-- name: GetSuppressedEmails :many
-- Returns emails that should never be contacted again.
SELECT DISTINCT "to" FROM email_log
WHERE workspace_id = $1
  AND (status = 'suppressed'
       OR bounce_reason LIKE 'hard bounce%'
       OR EXISTS (
           SELECT 1 FROM email_event e
           WHERE e.email_log_id = email_log.id
             AND e.event_type IN ('complained', 'unsubscribed')
       ));

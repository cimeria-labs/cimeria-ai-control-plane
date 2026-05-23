-- Email events: track opens, clicks, bounces, and replies from Resend webhooks.

CREATE TABLE IF NOT EXISTS email_event (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email_log_id UUID NOT NULL REFERENCES email_log(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL
        CHECK (event_type IN ('delivered', 'opened', 'clicked', 'replied', 'bounced', 'complained', 'unsubscribed', 'deferred', 'dropped')),
    ip_address INET,
    user_agent TEXT,
    link_url TEXT, -- original URL for click events
    city TEXT,
    country TEXT,
    metadata JSONB, -- extra fields from Resend webhook payload
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_email_event_log ON email_event(email_log_id);
CREATE INDEX IF NOT EXISTS idx_email_event_type ON email_event(event_type);
CREATE INDEX IF NOT EXISTS idx_email_event_created ON email_event(created_at DESC);

-- View for quick SDR queries: latest event per email log
CREATE OR REPLACE VIEW email_event_latest AS
SELECT DISTINCT ON (email_log_id) *
FROM email_event
ORDER BY email_log_id, created_at DESC;

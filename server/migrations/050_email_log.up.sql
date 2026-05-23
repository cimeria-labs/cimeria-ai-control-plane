-- Email log: track every email sent through the platform for analytics and follow-up.

CREATE TABLE IF NOT EXISTS email_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspace(id) ON DELETE CASCADE,
    issue_id UUID REFERENCES issue(id) ON DELETE SET NULL,
    sender_id UUID, -- member or agent who triggered the send
    sender_type TEXT CHECK (sender_type IN ('member', 'agent')),
    "to" TEXT NOT NULL,
    subject TEXT NOT NULL,
    body_preview TEXT, -- first 200 chars for quick reference
    resend_id TEXT, -- ID returned by Resend API
    email_type TEXT NOT NULL DEFAULT 'generic'
        CHECK (email_type IN ('cold_email', 'follow_up', 'proposal', 'reengagement', 'generic', 'verification', 'invitation')),
    status TEXT NOT NULL DEFAULT 'sent'
        CHECK (status IN ('sent', 'delivered', 'bounced', 'failed', 'suppressed')),
    tracking_enabled BOOLEAN NOT NULL DEFAULT true,
    opened_at TIMESTAMPTZ,
    clicked_at TIMESTAMPTZ,
    replied_at TIMESTAMPTZ,
    bounce_reason TEXT,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_email_log_workspace ON email_log(workspace_id);
CREATE INDEX IF NOT EXISTS idx_email_log_issue ON email_log(issue_id);
CREATE INDEX IF NOT EXISTS idx_email_log_resend ON email_log(resend_id) WHERE resend_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_email_log_to ON email_log("to");
CREATE INDEX IF NOT EXISTS idx_email_log_status ON email_log(status);

-- Leads are workspace-scoped sales records managed by the SDR workflow.

CREATE TABLE IF NOT EXISTS lead (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspace(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    company TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT 'manual',
    status TEXT NOT NULL DEFAULT 'captured'
        CHECK (status IN (
            'captured',
            'qualified',
            'rejected',
            'copy_ready',
            'strategy_ready',
            'email_sent',
            'nurturing',
            'hot',
            'handoff_human',
            'converted',
            'cancelled'
        )),
    score INTEGER NOT NULL DEFAULT 0 CHECK (score >= 0),
    dynamic_score INTEGER NOT NULL DEFAULT 0 CHECK (dynamic_score >= 0),
    assignee_type TEXT CHECK (assignee_type IN ('member', 'agent')),
    assignee_id UUID,
    pipeline_id UUID,
    state_machine_status TEXT NOT NULL DEFAULT 'captured',
    last_event TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, email)
);

ALTER TABLE email_log
    ADD COLUMN IF NOT EXISTS lead_id UUID REFERENCES lead(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_lead_workspace ON lead(workspace_id);
CREATE INDEX IF NOT EXISTS idx_lead_workspace_status ON lead(workspace_id, status);
CREATE INDEX IF NOT EXISTS idx_lead_workspace_email ON lead(workspace_id, lower(email));
CREATE INDEX IF NOT EXISTS idx_lead_updated ON lead(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_email_log_lead ON email_log(lead_id);

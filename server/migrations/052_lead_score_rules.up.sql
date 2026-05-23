-- Lead score rules: per-workspace configurable dynamic scoring weights.

CREATE TABLE IF NOT EXISTS lead_score_rule (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspace(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL
        CHECK (event_type IN ('opened', 'clicked', 'replied', 'forwarded', 'bounced', 'complained', 'unsubscribed')),
    weight INTEGER NOT NULL DEFAULT 0,
    max_per_email INTEGER, -- max times this event can count per email (e.g. open max 3)
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(workspace_id, event_type)
);

CREATE INDEX IF NOT EXISTS idx_lead_score_rule_workspace ON lead_score_rule(workspace_id);

-- Seed default rules for every existing workspace
INSERT INTO lead_score_rule (workspace_id, event_type, weight, max_per_email)
SELECT w.id, r.event_type, r.weight, r.max_per_email
FROM workspace w
CROSS JOIN (
    VALUES
        ('opened', 1, 3),
        ('clicked', 3, 2),
        ('replied', 5, 1),
        ('forwarded', 2, 1),
        ('bounced', -2, 1),
        ('complained', -5, 1),
        ('unsubscribed', -10, 1)
) AS r(event_type, weight, max_per_email)
ON CONFLICT (workspace_id, event_type) DO NOTHING;

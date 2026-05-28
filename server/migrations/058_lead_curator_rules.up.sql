-- Curator rules provide deterministic approve/reject/review recommendations.

CREATE TABLE IF NOT EXISTS lead_curator_rule (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspace(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('approve', 'reject', 'review')),
    field TEXT NOT NULL CHECK (field IN (
        'email',
        'company',
        'name',
        'title',
        'industry',
        'company_size',
        'icp_fit',
        'budget',
        'authority',
        'need',
        'timeline'
    )),
    operator TEXT NOT NULL CHECK (operator IN (
        'exists',
        'not_exists',
        'contains',
        'not_contains',
        'eq',
        'ne',
        'gt',
        'gte',
        'lt',
        'lte',
        'regex',
        'domain_in',
        'domain_not_in'
    )),
    value TEXT,
    priority INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    match_count INTEGER NOT NULL DEFAULT 0 CHECK (match_count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_lead_curator_rule_workspace ON lead_curator_rule(workspace_id);
CREATE INDEX IF NOT EXISTS idx_lead_curator_rule_workspace_active ON lead_curator_rule(workspace_id, is_active);
CREATE INDEX IF NOT EXISTS idx_lead_curator_rule_priority ON lead_curator_rule(priority DESC, created_at DESC);

-- Lead sources describe where candidates come from. Config must contain only non-secret filters.

CREATE TABLE IF NOT EXISTS lead_source (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspace(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    provider TEXT NOT NULL CHECK (provider IN (
        'manual',
        'csv',
        'api',
        'form',
        'apollo',
        'hunter',
        'linkedin',
        'referral',
        'website',
        'hubspot',
        'pipedrive'
    )),
    config JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    auto_approve BOOLEAN NOT NULL DEFAULT false,
    enrichment_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_lead_source_workspace ON lead_source(workspace_id);
CREATE INDEX IF NOT EXISTS idx_lead_source_workspace_provider ON lead_source(workspace_id, provider);

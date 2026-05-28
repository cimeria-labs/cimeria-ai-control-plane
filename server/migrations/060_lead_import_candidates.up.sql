-- Import candidates are preview/enrichment records that are not leads yet.
-- This lets Apollo search remain no-send and human-approved before lead creation.

CREATE TABLE IF NOT EXISTS lead_import_candidate (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspace(id) ON DELETE CASCADE,
    batch_id UUID NOT NULL REFERENCES lead_import_batch(id) ON DELETE CASCADE,
    provider TEXT NOT NULL DEFAULT 'apollo' CHECK (provider IN (
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
    external_id TEXT NOT NULL,
    email TEXT,
    email_status TEXT,
    name TEXT NOT NULL DEFAULT '',
    company TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL DEFAULT '',
    domain TEXT NOT NULL DEFAULT '',
    linkedin_url TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'preview' CHECK (status IN (
        'preview',
        'approved',
        'enriched',
        'imported',
        'duplicate',
        'rejected',
        'missing_email',
        'failed'
    )),
    score INTEGER NOT NULL DEFAULT 0 CHECK (score >= 0),
    payload JSONB NOT NULL DEFAULT '{}',
    enriched_payload JSONB NOT NULL DEFAULT '{}',
    error TEXT,
    lead_id UUID REFERENCES lead(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workspace_id, provider, external_id)
);

CREATE INDEX IF NOT EXISTS idx_lead_import_candidate_workspace ON lead_import_candidate(workspace_id);
CREATE INDEX IF NOT EXISTS idx_lead_import_candidate_batch ON lead_import_candidate(batch_id);
CREATE INDEX IF NOT EXISTS idx_lead_import_candidate_status ON lead_import_candidate(workspace_id, status);
CREATE INDEX IF NOT EXISTS idx_lead_import_candidate_email ON lead_import_candidate(workspace_id, lower(email));
CREATE INDEX IF NOT EXISTS idx_lead_import_candidate_lead ON lead_import_candidate(lead_id);

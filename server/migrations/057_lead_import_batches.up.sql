-- Import batches track preview/import lifecycle and preserve provider metadata.

CREATE TABLE IF NOT EXISTS lead_import_batch (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspace(id) ON DELETE CASCADE,
    source_id UUID REFERENCES lead_source(id) ON DELETE SET NULL,
    file_name TEXT,
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
    total_rows INTEGER NOT NULL DEFAULT 0 CHECK (total_rows >= 0),
    imported_count INTEGER NOT NULL DEFAULT 0 CHECK (imported_count >= 0),
    duplicate_count INTEGER NOT NULL DEFAULT 0 CHECK (duplicate_count >= 0),
    rejected_count INTEGER NOT NULL DEFAULT 0 CHECK (rejected_count >= 0),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN (
        'pending',
        'preview',
        'importing',
        'completed',
        'failed',
        'cancelled'
    )),
    error_log TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE lead
    ADD COLUMN IF NOT EXISTS import_batch_id UUID REFERENCES lead_import_batch(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_lead_import_batch_workspace ON lead_import_batch(workspace_id);
CREATE INDEX IF NOT EXISTS idx_lead_import_batch_workspace_status ON lead_import_batch(workspace_id, status);
CREATE INDEX IF NOT EXISTS idx_lead_import_batch_source ON lead_import_batch(source_id);
CREATE INDEX IF NOT EXISTS idx_lead_import_batch_provider ON lead_import_batch(provider);
CREATE INDEX IF NOT EXISTS idx_lead_import_batch_created ON lead_import_batch(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_lead_import_batch_lead ON lead(import_batch_id);

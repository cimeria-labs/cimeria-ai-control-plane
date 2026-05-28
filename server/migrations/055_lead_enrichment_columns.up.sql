-- Lead enrichment fields used by generated lead queries and SDR curation.

ALTER TABLE lead
    ADD COLUMN IF NOT EXISTS budget TEXT NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS authority TEXT NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS need TEXT NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS timeline TEXT NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS company_size TEXT NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS industry TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS pain_points TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS icp_fit TEXT NOT NULL DEFAULT 'unknown',
    ADD COLUMN IF NOT EXISTS lead_temperature TEXT NOT NULL DEFAULT 'cold',
    ADD COLUMN IF NOT EXISTS curated_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS curated_by UUID REFERENCES member(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_lead_workspace_icp_fit ON lead(workspace_id, icp_fit);
CREATE INDEX IF NOT EXISTS idx_lead_workspace_temperature ON lead(workspace_id, lead_temperature);
CREATE INDEX IF NOT EXISTS idx_lead_curated_by ON lead(curated_by);

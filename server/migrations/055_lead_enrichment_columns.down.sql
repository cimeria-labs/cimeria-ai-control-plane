DROP INDEX IF EXISTS idx_lead_curated_by;
DROP INDEX IF EXISTS idx_lead_workspace_temperature;
DROP INDEX IF EXISTS idx_lead_workspace_icp_fit;

ALTER TABLE lead
    DROP COLUMN IF EXISTS curated_by,
    DROP COLUMN IF EXISTS curated_at,
    DROP COLUMN IF EXISTS lead_temperature,
    DROP COLUMN IF EXISTS icp_fit,
    DROP COLUMN IF EXISTS pain_points,
    DROP COLUMN IF EXISTS industry,
    DROP COLUMN IF EXISTS company_size,
    DROP COLUMN IF EXISTS timeline,
    DROP COLUMN IF EXISTS need,
    DROP COLUMN IF EXISTS authority,
    DROP COLUMN IF EXISTS budget;

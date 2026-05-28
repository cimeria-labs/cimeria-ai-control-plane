DROP INDEX IF EXISTS idx_lead_import_batch_lead;
DROP INDEX IF EXISTS idx_lead_import_batch_created;
DROP INDEX IF EXISTS idx_lead_import_batch_provider;
DROP INDEX IF EXISTS idx_lead_import_batch_source;
DROP INDEX IF EXISTS idx_lead_import_batch_workspace_status;
DROP INDEX IF EXISTS idx_lead_import_batch_workspace;

ALTER TABLE lead DROP COLUMN IF EXISTS import_batch_id;

DROP TABLE IF EXISTS lead_import_batch;

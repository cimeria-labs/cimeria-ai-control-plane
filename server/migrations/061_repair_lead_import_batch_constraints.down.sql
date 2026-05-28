ALTER TABLE lead_import_batch
    DROP CONSTRAINT IF EXISTS lead_import_batch_status_check,
    ADD CONSTRAINT lead_import_batch_status_check CHECK (status IN (
        'pending',
        'processing',
        'completed',
        'failed'
    )) NOT VALID;

ALTER TABLE lead_import_batch
    DROP CONSTRAINT IF EXISTS lead_import_batch_provider_check,
    ADD CONSTRAINT lead_import_batch_provider_check CHECK (provider IN (
        'csv',
        'api',
        'form',
        'manual'
    )) NOT VALID;

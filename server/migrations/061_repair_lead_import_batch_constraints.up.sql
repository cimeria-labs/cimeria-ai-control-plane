-- Repair drift from early deployed lead_import_batch definitions.
-- Some VM databases had migration 057 marked applied while keeping the old
-- provider/status checks, which blocked Apollo preview batches.

ALTER TABLE lead_import_batch
    DROP CONSTRAINT IF EXISTS lead_import_batch_provider_check,
    ADD CONSTRAINT lead_import_batch_provider_check CHECK (provider IN (
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
    ));

ALTER TABLE lead_import_batch
    DROP CONSTRAINT IF EXISTS lead_import_batch_status_check,
    ADD CONSTRAINT lead_import_batch_status_check CHECK (status IN (
        'pending',
        'preview',
        'importing',
        'processing',
        'completed',
        'failed',
        'cancelled'
    ));

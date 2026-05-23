-- Index for fast webhook token lookups.
CREATE INDEX IF NOT EXISTS idx_autopilot_trigger_webhook_token ON autopilot_trigger(webhook_token)
    WHERE kind = 'webhook' AND enabled = true;
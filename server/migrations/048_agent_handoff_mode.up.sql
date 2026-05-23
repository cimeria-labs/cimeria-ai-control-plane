ALTER TABLE agent ADD COLUMN handoff_mode TEXT NOT NULL DEFAULT 'automatic';
-- Values: 'automatic' (default), 'manual'
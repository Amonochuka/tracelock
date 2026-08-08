ALTER TABLE access_events ADD COLUMN IF NOT EXISTS reason VARCHAR(50);
ALTER TABLE access_events ADD COLUMN IF NOT EXISTS device_id INT REFERENCES devices(id) ON DELETE SET NULL;
ALTER TABLE access_events ADD COLUMN IF NOT EXISTS entry_method VARCHAR(20) CHECK (entry_method IN ('fingerprint', 'face', 'iris', 'card', 'pin', 'api'));

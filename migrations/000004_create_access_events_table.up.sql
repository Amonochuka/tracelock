CREATE TABLE IF NOT EXISTS access_events (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    zone_id INT REFERENCES zones(id) ON DELETE SET NULL,
    device_id INT REFERENCES devices(id) ON DELETE SET NULL,
    entry_method VARCHAR(20) CHECK (entry_method IN ('fingerprint', 'face', 'iris', 'card', 'pin', 'api')),
    action VARCHAR(10) NOT NULL CHECK (action IN ('enter', 'exit')),
    status VARCHAR(10) NOT NULL CHECK (status IN ('allowed', 'denied')),
    reason VARCHAR(50),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    hash VARCHAR(64) NOT NULL,
    previous_hash VARCHAR(64)
);

CREATE INDEX IF NOT EXISTS idx_access_events_zone_ts ON access_events(zone_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_access_events_user_ts ON access_events(user_id, timestamp DESC);

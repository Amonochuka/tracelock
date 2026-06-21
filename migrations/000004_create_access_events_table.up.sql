CREATE TABLE IF NOT EXISTS access_events (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    zone_id INT REFERENCES zones(id),
    device_id INT REFERENCES devices(id),
    entry_method VARCHAR(20) CHECK (entry_method IN ('fingerprint', 'face', 'iris', 'card', 'pin', 'api')),
    action VARCHAR(10) NOT NULL CHECK (action IN ('enter', 'exit')),
    status VARCHAR(10) NOT NULL CHECK (status IN ('allowed', 'denied')),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    hash VARCHAR(64) NOT NULL,
    previous_hash VARCHAR(64)
);

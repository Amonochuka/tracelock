CREATE TABLE IF NOT EXISTS access_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    zone_id INTEGER,
    device_id INTEGER,
    entry_method TEXT CHECK (entry_method IN ('fingerprint', 'face', 'iris', 'card', 'pin', 'api')),
    action TEXT NOT NULL CHECK (action IN ('enter', 'exit')),
    status TEXT NOT NULL CHECK (status IN ('allowed', 'denied')),
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    hash TEXT NOT NULL,
    previous_hash TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (zone_id) REFERENCES zones(id),
    FOREIGN KEY (device_id) REFERENCES devices(id)
);

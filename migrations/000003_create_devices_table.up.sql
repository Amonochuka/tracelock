CREATE TABLE IF NOT EXISTS devices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    zone_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('fingerprint', 'face', 'iris', 'card', 'pin')),
    serial TEXT UNIQUE,
    active INTEGER NOT NULL DEFAULT 1,
    is_entry_point INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (zone_id) REFERENCES zones(id) ON DELETE CASCADE
);

-- Seed Data (Targeting specific unique constraint on serial)
INSERT INTO devices(zone_id, name, type, serial, is_entry_point) VALUES
(1, 'trinity', 'fingerprint', '1234abcd', 1),
(2, 'halo', 'iris', '5678efgh', 0),
(3, 'meta', 'face', '9012ijkl', 0),
(4, 'riswa', 'card', '4512mnop', 0)
ON CONFLICT(serial) DO NOTHING;

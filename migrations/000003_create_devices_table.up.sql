CREATE TABLE IF NOT EXISTS devices (
    id SERIAL PRIMARY KEY,
    zone_id INT NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('fingerprint', 'face', 'iris', 'card', 'pin')),
    serial VARCHAR(100) UNIQUE,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    is_entry_point BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seed Data (Targeting specific unique constraint on serial)
INSERT INTO devices(zone_id, name, type, serial, is_entry_point) VALUES
(1, 'trinity', 'fingerprint', '1234abcd', TRUE),
(2, 'halo', 'iris', '5678efgh', FALSE),
(3, 'meta', 'face', '9012ijkl', FALSE),
(4, 'riswa', 'card', '4512mnop', FALSE)
ON CONFLICT (serial) DO NOTHING;

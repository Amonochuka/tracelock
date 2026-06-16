CREATE TABLE IF NOT EXISTS zones (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    max_capacity INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Seed Data
INSERT INTO zones(name, description, max_capacity) VALUES
('Lobby', 'Main lobby', 50),
('Server Room', 'Restricted area', 5),
('Gym', 'Employee gym', 20),
('Boardroom', 'Meeting room', 20)
ON CONFLICT (name) DO NOTHING;

CREATE TABLE IF NOT EXISTS zones (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    max_capacity INT DEFAULT 0,
    requires_exit_scan BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seed Data
INSERT INTO zones(name, description, max_capacity, requires_exit_scan) VALUES
('Lobby', 'Main lobby', 50, FALSE),
('Server Room', 'Restricted area', 5, TRUE),
('Gym', 'Employee gym', 20, FALSE),
('Boardroom', 'Meeting room', 20, FALSE)
ON CONFLICT (name) DO NOTHING;

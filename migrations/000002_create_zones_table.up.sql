CREATE TABLE IF NOT EXISTS zones (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    max_capacity INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seed Data
INSERT INTO zones(name, description, max_capacity) VALUES
('Lobby', 'Main lobby', 50),
('Server Room', 'Restricted area', 5),
('Gym', 'Employee gym', 20),
('Boardroom', 'Meeting room', 20)
ON CONFLICT (name) DO NOTHING;

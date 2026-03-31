CREATE TABLE IF NOT EXISTS users(
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS zones(
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    max_capacity INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS access_events(
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    zone_id INT REFERENCES zones(id),
    action VARCHAR(10) NOT NULL CHECK (action IN ('enter', 'exit')),
    status VARCHAR(10) NOT NULL CHECK (status IN ('allowed', 'denied')),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    hash VARCHAR(64) NOT NULL,
    previous_hash VARCHAR(64)
);

CREATE TABLE IF NOT EXISTS active_sessions (
    user_id INT NOT NULL REFERENCES users(id),
    zone_id INT NOT NULL REFERENCES zones(id),
    entered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(user_id, zone_id)
);

INSERT INTO zones(name, description, max_capacity) VALUES
('Lobby', 'Main lobby', 50),
('Server Room', 'Restricted area', 5),
('Gym', 'Employee gym', 20),
('Boardroom', 'Meeting room', 20)
ON CONFLICT DO NOTHING;
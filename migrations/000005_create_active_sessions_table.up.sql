CREATE TABLE IF NOT EXISTS active_sessions (
    user_id INT NOT NULL REFERENCES users(id),
    zone_id INT NOT NULL REFERENCES zones(id),
    entered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(user_id, zone_id)
);

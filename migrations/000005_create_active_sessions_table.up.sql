CREATE TABLE IF NOT EXISTS active_sessions (
    user_id INTEGER NOT NULL,
    zone_id INTEGER NOT NULL,
    entered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(user_id, zone_id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (zone_id) REFERENCES zones(id)
);

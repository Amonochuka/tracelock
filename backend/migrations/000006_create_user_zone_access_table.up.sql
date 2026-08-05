CREATE TABLE IF NOT EXISTS user_zone_access (
    user_id    INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    zone_id    INT NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
    granted_by INT REFERENCES users(id),
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(user_id, zone_id)
);

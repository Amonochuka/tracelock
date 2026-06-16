CREATE TABLE IF NOT EXISTS biometric_credentials (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         INTEGER NOT NULL,
    entry_method    TEXT NOT NULL CHECK (entry_method IN ('fingerprint', 'face', 'iris', 'card', 'pin')),
    credential_hash TEXT NOT NULL,
    enrolled_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
    revoked         INTEGER NOT NULL DEFAULT 0,
    UNIQUE (user_id, entry_method),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

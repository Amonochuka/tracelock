CREATE TABLE IF NOT EXISTS biometric_credentials (
    id              SERIAL PRIMARY KEY,
    user_id         INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    entry_method    VARCHAR(20) NOT NULL CHECK (entry_method IN ('fingerprint', 'face', 'iris', 'card', 'pin')),
    credential_hash VARCHAR(255) NOT NULL,
    enrolled_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    revoked         BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE (user_id, entry_method)
);

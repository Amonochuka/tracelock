-- Store lock expiry as TIMESTAMPTZ so a mismatch between the database
-- session timezone and the app server timezone can never stretch a
-- 15-minute lock into hours. Existing naive values are interpreted as UTC.
ALTER TABLE users
    ALTER COLUMN locked_until TYPE TIMESTAMPTZ
    USING locked_until AT TIME ZONE 'UTC';

ALTER TABLE users
    ALTER COLUMN locked_until TYPE TIMESTAMP
    USING locked_until AT TIME ZONE 'UTC';

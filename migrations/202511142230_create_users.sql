-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    provider TEXT NOT NULL,
    provider_id TEXT NOT NULL,
    email TEXT,
    name TEXT,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    UNIQUE (provider, provider_id)
);

-- +goose Down
DROP TABLE IF EXISTS users;
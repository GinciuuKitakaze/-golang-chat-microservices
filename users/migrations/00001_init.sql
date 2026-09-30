-- +goose Up
CREATE TABLE users (
    id         UUID PRIMARY KEY,
    email      VARCHAR(255) UNIQUE,
    phone      VARCHAR(255) UNIQUE,
    name VARCHAR(255) NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    version    BIGINT       NOT NULL DEFAULT 1,
    created_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    CHECK (email IS NOT NULL OR phone IS NOT NULL)
);

-- +goose Down
DROP TABLE IF EXISTS users;
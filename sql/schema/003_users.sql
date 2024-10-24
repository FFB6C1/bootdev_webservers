-- +goose Up
ALTER TABLE users
ADD COLUMN hashed_password TEXT NOT NULL
CONSTRAINT pw_default DEFAULT 'unset';

-- +goose Down
ALTER TABLE users
DROP COLUMN hashed_password;
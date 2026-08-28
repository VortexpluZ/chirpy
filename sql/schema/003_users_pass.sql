-- +goose Up
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS hashed_password TEXT NOT NULL DEFAULT 'unset';

ALTER TABLE users
  ALTER COLUMN hashed_password DROP DEFAULT;

-- +goose Down
ALTER TABLE users
  DROP COLUMN IF EXISTS hashed_password;
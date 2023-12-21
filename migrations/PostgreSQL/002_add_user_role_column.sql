-- +goose Up
ALTER TABLE user_profiles
ADD COLUMN user_role INTEGER;

-- +goose Down
ALTER TABLE user_profiles
DROP COLUMN IF EXISTS user_role;
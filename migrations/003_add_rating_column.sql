-- +goose Up
ALTER TABLE user_profiles
ADD COLUMN rating INTEGER;

-- +goose Down
ALTER TABLE user_profiles
DROP COLUMN IF EXISTS rating;
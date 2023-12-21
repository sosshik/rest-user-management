-- +goose Up

CREATE TABLE IF NOT EXISTS rating.emotes (
    from_oid UUID,
    to_oid UUID,
    emoji_id Int,
    voted_at DateTime
) ENGINE = SummingMergeTree()
ORDER BY ();

-- +goose Down

DROP TABLE IF EXISTS rating.emotes;
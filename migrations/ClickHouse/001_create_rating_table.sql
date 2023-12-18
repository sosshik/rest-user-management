-- +goose Up

CREATE TABLE IF NOT EXISTS emotes (
    id Int PRIMARY KEY,
    from_oid UUID,
    to_oid UUID,
    emoji_id Int,
    voted_at DateTime
) ENGINE = SummingMergeTree()
ORDER BY (id);

-- +goose Down

DROP TABLE IF EXISTS;
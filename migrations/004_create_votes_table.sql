-- +goose Up
CREATE TABLE IF NOT EXISTS votes (
    id SERIAL PRIMARY KEY,
    from_oid UUID NOT NULL,
    to_oid UUID NOT NULL,
    value INT NOT NULL,
    voted_at TIMESTAMP NOT NULL
);

-- +goose Down

DROP TABLE votes;
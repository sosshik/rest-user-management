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

INSERT INTO rating.emotes (to_oid, from_oid, em_type, votes) VALUES
    ('8767614f-4853-47d8-854f-e93c66359ef9', '2b20f3f7-0185-4457-921e-53a9d233e166', 0, 1),
    ('8767614f-4853-47d8-854f-e93c66359ef9', '2546e9f3-609f-461f-9ecf-08901b7fe897', 0, 1),
    ('2b20f3f7-0185-4457-921e-53a9d233e166', '2546e9f3-609f-461f-9ecf-08901b7fe897', 1, 1),
    ('2546e9f3-609f-461f-9ecf-08901b7fe897', '8767614f-4853-47d8-854f-e93c66359ef9', 0, 1);

CREATE TABLE IF NOT EXISTS rating.emotes (
to_oid UUID,
from_oid UUID,
em_type Int,
votes Int
) ENGINE = SummingMergeTree()
ORDER BY ();

CREATE TABLE IF NOT EXISTS rating.emotes (
    from_oid UUID,
    to_oid UUID,
    emoji_id Int,
    year Int,
    month Int,
    day int,
    hour int, 
    minute int
) ENGINE = SummingMergeTree()
ORDER BY ();
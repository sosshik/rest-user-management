package rating

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"git.foxminded.ua/foxstudent106264/task-3.5/cmd/internal/domain"
	"git.foxminded.ua/foxstudent106264/task-3.5/pkg/config"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
)

type ClickHouse struct {
	conn driver.Conn
}

func NewClickHouse(cfg *config.Config) (*ClickHouse, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.CH.Addr},
		Auth: clickhouse.Auth{
			Database: cfg.CH.DB,
			Username: cfg.CH.User,
			Password: cfg.CH.Pass,
		},
	})
	if err != nil {
		return &ClickHouse{}, err
	}

	return &ClickHouse{conn: conn}, nil
}

func (c *ClickHouse) RateProfile(vote domain.VoteDTO) error {
	err := c.conn.Exec(context.Background(), `
	INSERT INTO rating.votes (from_oid, to_oid, emoji_id, voted_at)
	VALUES ($1,$2,$3,$4);
	`, vote.FromOID, vote.ToOID, vote.EmojiId, vote.VotedAt)
	if err != nil {
		return fmt.Errorf("RateProfile: unable to execute query to DB: %w", err)
	}

	return nil
}

func (c *ClickHouse) GetVote(vote domain.VoteDTO) (domain.VoteDTO, bool, error) {
	var dbVote domain.VoteDTO
	err := c.conn.QueryRow(context.Background(), `
		SELECT from_oid, to_oid, emoji_id, voted_at FROM rating.votes
		WHERE from_oid = $1 AND to_oid = $2;
	`, vote.FromOID, vote.ToOID).Scan(&dbVote.FromOID, &dbVote.ToOID, &dbVote.EmojiId, &dbVote.VotedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.VoteDTO{}, false, err
		}
		return domain.VoteDTO{}, false, fmt.Errorf("GetVote: unable to execute query to DB: %w", err)
	}
	return dbVote, true, nil
}

func (c *ClickHouse) LastVotedAt(vote domain.VoteDTO) (time.Time, error) {
	var lastVoted time.Time
	err := c.conn.QueryRow(context.Background(), `
		SELECT voted_at FROM rating.votes
		WHERE from_oid = $1
		ORDER BY voted_at DESC 
		LIMIT 1;
	`, vote.FromOID).Scan(&lastVoted)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("LastVotedAt: unable to execute query to DB: %w", err)
	}
	return lastVoted, nil
}

func (c *ClickHouse) UpdateProfileRating(vote domain.VoteDTO, oldValue int32) error {

	err := c.conn.Exec(context.Background(), `
		ALTER TABLE rating.votes
		UPDATE emoji_id = $1, voted_at = $2
		WHERE from_oid = $3 AND to_oid = $4;
	`, vote.EmojiId, vote.VotedAt, vote.FromOID, vote.ToOID)
	if err != nil {
		return fmt.Errorf("UpdateProfileRating: unable to execute query to update votes: %w", err)
	}

	return nil
}

func (c *ClickHouse) GetRating(userId uuid.UUID) (int, error) {
	var rating uint64
	err := c.conn.QueryRow(context.Background(), `
	SELECT COUNT(*)
    FROM rating.votes
	WHERE to_oid = $1;
	`, userId).Scan(&rating)
	if err != nil {
		return 0, fmt.Errorf("GetRating: unable to execute query to DB: %w", err)
	}
	return int(rating), nil
}

func (c *ClickHouse) GetRatingSeparately(userId uuid.UUID) (string, error) {
	ratings := map[int]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0}
	for i := 0; i < len(ratings); i++ {
		var rating uint64
		err := c.conn.QueryRow(context.Background(), `
		SELECT COUNT(*)
		FROM rating.votes
		WHERE to_oid = $1 AND emoji_id = $2;
		`, userId, i+1).Scan(&rating)
		if err != nil {
			return "", fmt.Errorf("GetRatingSeparately: unable to execute query to DB: %w", err)
		}
		ratings[i+1] = int(rating)
	}

	ratingStr := ""

	for i := 0; i < len(ratings); i++ {
		ratingStr += fmt.Sprintf("%s:%d; ", emojiStr[i+1], ratings[i+1])
	}

	return ratingStr, nil
}

// "emoji_id: %d, count: %d; "

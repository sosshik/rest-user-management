package rating

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"github.com/sosshik/rest-user-management/cmd/internal/domain"
	"github.com/sosshik/rest-user-management/pkg/config"
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
	INSERT INTO rating.emotes (from_oid, to_oid, emoji_id, voted_at)
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
		SELECT from_oid, to_oid, emoji_id, voted_at FROM rating.emotes
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
		SELECT voted_at FROM rating.emotes
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
		ALTER TABLE rating.emotes
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
    FROM rating.emotes
	WHERE to_oid = $1;
	`, userId).Scan(&rating)
	if err != nil {
		return 0, fmt.Errorf("GetRating: unable to execute query to DB: %w", err)
	}
	return int(rating), nil
}

func (c *ClickHouse) GetRatingForList(oids []uuid.UUID) (map[uuid.UUID]int, error) {
	query := strings.Builder{}
	query.WriteString("SELECT to_oid, COUNT(*) FROM rating.emotes WHERE to_oid IN (")
	for i, oid := range oids {
		if i > 0 {
			query.WriteString(",")
		}

		query.WriteString(fmt.Sprintf("'%s',", oid.String()))
	}
	query.WriteString(") GROUP BY to_oid;")

	ratings := make(map[uuid.UUID]int)

	rows, err := c.conn.Query(context.Background(), query.String())
	if err != nil {
		return map[uuid.UUID]int{}, fmt.Errorf("GetRatingForList: unable to execute query to DB: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var oid uuid.UUID
		var rating uint64
		err := rows.Scan(&oid, &rating)
		if err != nil {
			return map[uuid.UUID]int{}, fmt.Errorf("GetRatingForList: scan the row: %w", err)
		}

		ratings[oid] = int(rating)
	}

	return ratings, nil

}

func (c *ClickHouse) GetRatingSeparately(userId uuid.UUID) (string, error) {
	ratingBuilder := strings.Builder{}
	for emojiId, emoji := range emojiStr {
		var rating uint64
		err := c.conn.QueryRow(context.Background(), `
		SELECT COUNT(*)
		FROM rating.emotes
		WHERE to_oid = $1 AND emoji_id = $2;
		`, userId, emojiId).Scan(&rating)
		if err != nil {
			return "", fmt.Errorf("GetRatingSeparately: unable to execute query to DB: %w", err)
		}
		ratingBuilder.WriteString(fmt.Sprintf("%s:%d; ", emoji, rating))
	}

	return ratingBuilder.String(), nil

}

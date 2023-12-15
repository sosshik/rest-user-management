package database

import (
	"time"

	"git.foxminded.ua/foxstudent106264/task-3.5/cmd/internal/domain"
	"github.com/google/uuid"
)

type UserProfile struct {
	ID        int          `json:"id"`
	OID       uuid.UUID    `json:"oid"`
	Nickname  string       `json:"nickname"`
	FirstName string       `json:"first_name"`
	LastName  string       `json:"last_name"`
	Password  string       `json:"password,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	State     domain.State `json:"state"`
	Role      domain.Role  `json:"user_role"`
	Rating    int          `json:"rating"`
}

type Vote struct {
	ID      int       `json:"id"`
	FromOID uuid.UUID `json:"from_oid"`
	ToOID   uuid.UUID `json:"to_oid"`
	Value   int       `json:"value"`
	VotedAt time.Time `json:"voted_at"`
}

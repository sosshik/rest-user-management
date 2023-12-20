package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role int

type State int

const (
	Usr Role = iota + 1
	Moderator
	Admin
)

const (
	Deleted State = iota - 1
	Banned
	Active
)

type UserProfileManager interface {
	CreateUserProfile(user UserProfileDTO) error
	UpdateUserProfile(user UserProfileDTO, oid uuid.UUID) error
	UpdatePassword(newPass string, oid uuid.UUID) error
	GetUserById(userID uuid.UUID) (UserProfileDTO, error)
	GetUserForToken(nickname string) (UserProfileDTO, error)
	GetUsersList(pageSize int, offset int) ([]UserProfileDTO, error)
	DeleteUser(oid uuid.UUID) error
	GetPassword(nickname string) (string, error)
	GetUsersCount() (int, error)
	GetUserState(oid uuid.UUID) (int, error)
}

type StatsManager interface {
	RateProfile(vote VoteDTO) error
	GetVote(vote VoteDTO) (VoteDTO, bool, error)
	LastVotedAt(vote VoteDTO) (time.Time, error)
	UpdateProfileRating(vote VoteDTO, oldRating int32) error
	GetRating(userId uuid.UUID) (int, error)
	GetRatingSeparately(userId uuid.UUID) (string, error)
	GetRatingForList(oids []uuid.UUID) (map[uuid.UUID]int, error)
}

type DomainInterface interface {
	UserProfileManager
	StatsManager
}

type CacheInterface interface {
	Set(key string, value interface{}) error
	GetUser(key string) (UserProfileDTO, error)
	GetUsersList(key string) (Pagination[UserProfileDTO], error)
	MakeKey(pageSize int, offset int) string
}

type UserProfileDTO struct {
	OID       uuid.UUID `json:"oid"`
	Nickname  string    `json:"nickname"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	State     State     `json:"state"`
	Role      Role      `json:"user_role"`
	Rating    int       `json:"rating"`
}

type GetProfileDTO struct {
	OID       uuid.UUID `json:"oid"`
	Nickname  string    `json:"nickname"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	State     State     `json:"state"`
	Role      Role      `json:"user_role"`
	Rating    string    `json:"rating"`
}

type VoteDTO struct {
	FromOID uuid.UUID `json:"from_oid"`
	ToOID   uuid.UUID `json:"oid"`
	EmojiId int32     `json:"emoji"`
	VotedAt time.Time `json:"voted_at"`
}

type Pagination[T any] struct {
	TotalItems  int `json:"total_items"`
	CurrentPage int `json:"current_page"`
	Users       []T `json:"users"`
}

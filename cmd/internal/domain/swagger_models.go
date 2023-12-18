package domain

import (
	"time"

	"github.com/google/uuid"
)

type CreateUserReq struct {
	Nickname  string `json:"nickname"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
}

type CreateUserResp struct {
	OID     uuid.UUID `json:"oid"`
	Message string    `json:"message"`
}

type LoginReq struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

type LoginResp struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

type UpdateUserReq struct {
	Nickname  string `json:"nickname"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type MessageResp struct {
	Message string `json:"message"`
}

type UpdatePasswordReq struct {
	Password string `json:"password"`
}

type GetUserResp struct {
	OID       uuid.UUID `json:"oid"`
	Nickname  string    `json:"nickname"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	State     int       `json:"state"`
}

type GetUserListResp struct {
	TotalUsers int           `json:"total_users"`
	Page       int           `json:"page"`
	Users      []GetUserResp `json:"users"`
}

type ErrorResp struct {
	Error string `json:"error"`
}

type VoteReq struct {
	OID   uuid.UUID `json:"oid"`
	Emoji int       `json:"emoji"`
}

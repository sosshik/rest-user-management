package api

type Pagination[T any] struct {
	TotalItems  int `json:"total_items"`
	CurrentPage int `json:"current_page"`
	Users       []T `json:"users"`
}

const defaultPageSize = 10

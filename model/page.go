package model

type Page[T any] struct {
	Total int `json:"total"`
	List  []T `json:"list"`
}

func ToPage[T any](total int, list []T) *Page[T] {
	return &Page[T]{Total: total, List: list}
}

package model

// Page represents a paginated result set for type T.
type Page[T any] struct {
	Total int `json:"total"` // Total number of items
	List  []T `json:"list"`  // Items on the current page
}

// ToPage creates a new Page from the given total count and item list.
func ToPage[T any](total int, list []T) *Page[T] {
	return &Page[T]{Total: total, List: list}
}

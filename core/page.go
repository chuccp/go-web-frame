package core

type Page[T IEntry] struct {
	Total int `json:"total"`
	List  []T `json:"list"`
}

func ToPage[T IEntry](total int, list []T) *Page[T] {
	return &Page[T]{Total: total, List: list}
}

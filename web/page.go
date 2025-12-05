package web

type Page struct {
	PageNo   int
	PageSize int
	LastId   int
}
type PageAble[T any] struct {
	Total int64 `json:"total"`
	List  any   `json:"list"`
}

func ToPage[T any](total int64, list any) *PageAble[T] {
	return &PageAble[T]{
		Total: total,
		List:  list,
	}
}

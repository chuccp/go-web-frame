package util

// Page represents pagination parameters for list queries.
type Page struct {
	PageNo   int // Current page number, 1-based
	PageSize int // Number of items per page
	LastId   int // Last seen ID for cursor-based pagination
}

func DefaultPage(page *Page) *Page {
	if page == nil {
		page = &Page{
			PageNo:   DefaultPageNo,
			PageSize: DefaultPageSize,
		}
		return page
	}
	if page.PageNo < 1 {
		page.PageNo = DefaultPageNo
	}
	if page.PageSize < 1 {
		page.PageSize = DefaultPageSize
	}
	return page
}

const DefaultPageSize = 10
const DefaultPageNo = 1

// PageAble is a paginated response wrapper containing total count and item list.
type PageAble[T any] struct {
	Total int64 `json:"total"` // Total number of items
	List  []T   `json:"list"`  // Items on the current page
}

// ToPage creates a new PageAble from the given total count and item list.
func ToPage[T any](total int64, list []T) *PageAble[T] {
	return &PageAble[T]{
		Total: total,
		List:  list,
	}
}

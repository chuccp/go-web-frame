package web2

type FilterChain interface {
	// Next executes the next filter or the final handler in the chain.
	// Returns (any, error): the result and any error.
	Next() (any, error)
}
type Filter interface {
	// Handle processes the request.
	// fc: the filter chain, used to invoke the next filter
	// request: the HTTP request object
	// Returns (any, error): the result and any error
	Handle(filterChain FilterChain, request *Request) (any, error)
}

type mockFilterChain struct {
	index     int
	request   *Request
	filters   []Filter
	converter Converter
	handler   HandlerFunc
}

func newMockFilterChain(request *Request, converter Converter, filters []Filter, handler HandlerFunc) *mockFilterChain {
	return &mockFilterChain{
		filters:   filters,
		index:     -1,
		request:   request,
		handler:   handler,
		converter: converter,
	}
}

func (c *mockFilterChain) Next() (any, error) {
	if c.index < len(c.filters)-1 {
		c.index++
		return c.filters[c.index].Handle(c, c.request)
	}
	return c.handler(c.request)
}

func (c *mockFilterChain) next() {
	if c.converter != nil {
		c.converter.Request(c, c.request)
	} else {
		defaultConverter.Request(c, c.request)
	}
}

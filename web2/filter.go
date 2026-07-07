package web2

import (
	"reflect"
	"runtime"
)

// HandlerChain is a chain of handler functions.
type HandlerChain []HandlerFunc

// GetFuncName returns the name of the last handler function in the chain.
func (c HandlerChain) GetFuncName() string {
	return runtime.FuncForPC(reflect.ValueOf(c.Last()).Pointer()).Name()
}

// Last returns the last handler in the chain, or nil if the chain is empty.
func (c HandlerChain) Last() HandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

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

type filterChain struct {
	index     int
	request   *Request
	filters   []Filter
	converter Converter
	handler   HandlerFunc
}

func newFilterChain(request *Request, converter Converter, filters []Filter, handler HandlerFunc) *filterChain {
	return &filterChain{
		filters:   filters,
		index:     -1,
		request:   request,
		handler:   handler,
		converter: converter,
	}
}

func (c *filterChain) Next() (any, error) {
	if c.index < len(c.filters)-1 {
		c.index++
		return c.filters[c.index].Handle(c, c.request)
	}
	return c.handler(c.request)
}

func (c *filterChain) next() {
	if c.converter != nil {
		c.converter.Request(c, c.request)
	} else {
		defaultConverter.Request(c, c.request)
	}
}

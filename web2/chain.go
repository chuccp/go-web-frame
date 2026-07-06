package web2

import (
	"reflect"
	"runtime"
)

type HandlersChain []HandlerFunc

// GetFuncName returns the name of the last handler function in the chain.
func (c HandlersChain) GetFuncName() string {
	return runtime.FuncForPC(reflect.ValueOf(c.Last()).Pointer()).Name()
}

// Last returns the last handler in the chain, or nil if the chain is empty.
func (c HandlersChain) Last() HandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

func Of(handlerInfo *HandlerInfo) HandlersChain {
	return handlerInfo.handlers
}

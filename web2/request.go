package web2

import (
	"github.com/gin-gonic/gin"
)

type JsonObject KV

type HandlerFunc func(*Request) (any, error)

type Request struct {
	c           *gin.Context
	jsonBody    *JsonObject
	handlerMeta *HandlerMeta
	response    Response
}

func (r *Request) Response() Response {
	return r.response

}

func request(ctx *gin.Context) *Request {

	return &Request{
		c:        ctx,
		response: newResponse(ctx),
	}
}

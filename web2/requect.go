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
}

package core

type IService interface {
	Init(context *Context)
	Name() string
}

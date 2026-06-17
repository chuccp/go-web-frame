package auth

import (
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
)

type Authentication[U any] interface {
	core.IService
	SignIn(user any, request *web.Request) (any, error)
	SignOut(request *web.Request) (any, error)
	User(request *web.Request) (U, error)
}

const (
	LoginKey = "login"
)

var NoLogin = &NoLoginError{}

type NoLoginError struct {
	error
}

func (e *NoLoginError) Error() string {
	return "no login"
}
func WithLogin() web.MetaOption {
	return web.WithKey(LoginKey)
}

type AuthenticationFilter[U any] struct {
	ctx            *core.Context
	authentication Authentication[U]
}

func (s *AuthenticationFilter[U]) Init(ctx *core.Context) error {
	s.ctx = ctx
	if s.authentication != nil {
		err := s.authentication.Init(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *AuthenticationFilter[U]) SignIn(user any, request *web.Request) (any, error) {
	return s.authentication.SignIn(user, request)
}
func (s *AuthenticationFilter[U]) SignOut(request *web.Request) (any, error) {

	return s.authentication.SignOut(request)
}
func (s *AuthenticationFilter[U]) User(request *web.Request) (U, error) {
	return s.authentication.User(request)
}
func (s *AuthenticationFilter[U]) Handle(filterChain web.FilterChain, request *web.Request) (any, error) {
	if request.HandlerMeta().Has(LoginKey) {
		_, err := s.authentication.User(request)
		if err != nil {
			return nil, err
		}
	}
	return filterChain.Next()
}

func User[T any](r *AuthenticationFilter[T], request *web.Request) (T, error) {
	u, err := r.User(request)
	if err != nil {
		var u T
		return u, err
	}
	return u, nil

}

func NewAuthenticationFilter[T any](authentication Authentication[T]) *AuthenticationFilter[T] {
	return &AuthenticationFilter[T]{
		authentication: authentication,
	}
}

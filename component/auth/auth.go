package auth

import (
	"reflect"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
)

type Authentication interface {
	core.IService
	SignIn(user any, request *web.Request) (any, error)
	SignOut(request *web.Request) (any, error)
	User(request *web.Request) (any, error)
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

type AuthenticationFilter struct {
	ctx            *core.Context
	authentication Authentication
}

func (s *AuthenticationFilter) Init(ctx *core.Context) error {
	s.ctx = ctx
	if s.authentication != nil {
		err := s.authentication.Init(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *AuthenticationFilter) SignIn(user any, request *web.Request) (any, error) {
	return s.authentication.SignIn(user, request)
}
func (s *AuthenticationFilter) SignOut(request *web.Request) (any, error) {

	return s.authentication.SignOut(request)
}
func (s *AuthenticationFilter) User(request *web.Request) (any, error) {
	return s.authentication.User(request)
}
func (s *AuthenticationFilter) Handle(filterChain web.FilterChain, request *web.Request) (any, error) {
	if request.HandlerMeta().Has(LoginKey) {
		user, err := s.authentication.User(request)
		if err != nil || user == nil {
			return web.Unauthorized("", err), nil
		}
	}
	return filterChain.Next()
}

func User[T any](r *AuthenticationFilter, request *web.Request) (T, error) {
	u, err := r.User(request)
	if err != nil {
		var u T
		return u, err
	}
	v, ok := u.(T)
	if !ok {
		return v, errors.New("Type conversion error. Please check if it is a pointer type." + reflect.TypeOf(u).Name())
	}
	return u.(T), nil

}

func NewAuthenticationFilter(authentication Authentication) *AuthenticationFilter {
	return &AuthenticationFilter{
		authentication: authentication,
	}
}

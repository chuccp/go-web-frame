package auth

import (
	"reflect"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
)

type Authentication interface {
	SignIn(user any, request *web.HttpContext) (any, error)
	SignOut(request *web.HttpContext) (any, error)
	User(request *web.HttpContext) (any, error)
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
	return nil
}
func (s *AuthenticationFilter) SignIn(user any, request *web.HttpContext) (any, error) {
	return s.authentication.SignIn(user, request)
}
func (s *AuthenticationFilter) SignOut(request *web.HttpContext) (any, error) {
	return s.authentication.SignOut(request)
}
func (s *AuthenticationFilter) User(request *web.HttpContext) (any, error) {
	return s.authentication.User(request)
}
func (s *AuthenticationFilter) Handle(request *web.HttpContext) {
	if request.HandlerMeta().Has(LoginKey) {
		response := request.Response()
		user, err := s.authentication.User(request)
		if err != nil || user == nil {
			response.AbortWithMessage(web.Unauthorized("", err))
			return
		}
	}
	request.Next()
}

func User[T any](r *AuthenticationFilter, request *web.HttpContext) (T, error) {
	u, err := r.User(request)
	if err != nil {
		return u.(T), err
	}
	v, ok := u.(T)
	if !ok {
		return v, errors.New("Type conversion error. Please check if it is a pointer type." + reflect.TypeOf(u).Name())
	}
	return u.(T), err

}

func NewAuthenticationFilter(authentication Authentication) *AuthenticationFilter {
	return &AuthenticationFilter{
		authentication: authentication,
	}
}

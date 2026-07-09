package auth

import (
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
)

// Authentication defines the interface for authentication services.
type Authentication[U any] interface {
	core.IService
	SignIn(value any, request *web.Request) (any, error)
	SignOut(request *web.Request) (any, error)
	User(request *web.Request) (U, error)
}

const (
	LoginKey = "login"
)

// NoLogin is the sentinel error returned when a user is not authenticated.
var NoLogin = web.NewForbidden().WithDetail("Not logged in")

// WithLogin returns a MetaOption that marks a route as requiring authentication.
func WithLogin() web.MetaOption {
	return web.WithKey(LoginKey)
}

// AuthenticationFilter is a filter that enforces authentication on protected routes.
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
	if request.HasMeta(WithLogin()) {
		_, err := s.authentication.User(request)
		if err != nil {
			return nil, web.NewForbidden().WithError(err)
		}
	}
	return filterChain.Next()
}

// User retrieves the authenticated user from the filter, cast to type T.
func User[T any](r *AuthenticationFilter[T], request *web.Request) (T, error) {
	u, err := r.User(request)
	if err != nil {
		var u T
		return u, err
	}
	return u, nil

}

// NewAuthenticationFilter creates a new authentication filter with the given provider.
func NewAuthenticationFilter[T any](authentication Authentication[T]) *AuthenticationFilter[T] {
	return &AuthenticationFilter[T]{
		authentication: authentication,
	}
}

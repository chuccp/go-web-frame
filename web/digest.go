package web

import "errors"

var NoLogin = &NoLoginError{}

type NoLoginError struct {
	error
}

func (e *NoLoginError) Error() string {
	return "no login"
}

type Authentication interface {
	SignIn(user any, request *Request) (any, error)
	SignOut(request *Request) (any, error)
	User(request *Request) (any, error)
	NewUser() any
}

type DigestAuth struct {
	authentication Authentication
}

func (digestAuth *DigestAuth) Authentication() Authentication {
	return digestAuth.authentication
}

func (digestAuth *DigestAuth) SignIn(user any, request *Request) (any, error) {
	if digestAuth.authentication != nil {
		return digestAuth.authentication.SignIn(user, request)
	}
	return nil, errors.New("secretProvider is nil")
}
func (digestAuth *DigestAuth) User(request *Request) (any, error) {
	if digestAuth.authentication != nil {
		return digestAuth.authentication.User(request)
	}
	return nil, errors.New("secretProvider is nil")
}

func (digestAuth *DigestAuth) SignOut(r *Request) (any, error) {
	return digestAuth.authentication.SignOut(r)

}

func NewDigestAuth(authentication Authentication) *DigestAuth {
	return &DigestAuth{authentication: authentication}
}

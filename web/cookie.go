// Package web: cookie helper for reading and writing HTTP cookies.
package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CookieOption is a functional option for configuring cookie attributes.
type CookieOption func(*http.Cookie)

// WithSecure sets the Secure flag (only send over HTTPS).
func WithSecure(secure bool) CookieOption {
	return func(c *http.Cookie) { c.Secure = secure }
}

// WithHttpOnly sets the HttpOnly flag (not accessible via JavaScript).
func WithHttpOnly(httpOnly bool) CookieOption {
	return func(c *http.Cookie) { c.HttpOnly = httpOnly }
}

// WithSameSite sets the SameSite attribute.
func WithSameSite(sameSite http.SameSite) CookieOption {
	return func(c *http.Cookie) { c.SameSite = sameSite }
}

// WithDomain sets the cookie domain.
func WithDomain(domain string) CookieOption {
	return func(c *http.Cookie) { c.Domain = domain }
}

// WithPath sets the cookie path.
func WithPath(path string) CookieOption {
	return func(c *http.Cookie) { c.Path = path }
}

// Cookie provides helper methods for reading, writing, and deleting HTTP cookies.
type Cookie struct {
	c *gin.Context
}

// Get returns the value of a cookie by key, or empty string if not found.
func (c *Cookie) Get(key string) string {
	cookie, err := c.c.Cookie(key)
	if err != nil {
		return ""
	}
	return cookie
}

// Set sets a cookie with optional custom attributes.
// Defaults: MaxAge=1 year, Path="/", HttpOnly=true, SameSite=Lax.
//
//	cookie.Set("token", value)                              // defaults
//	cookie.Set("token", value, WithSecure(true))            // HTTPS only
//	cookie.Set("token", value, WithSecure(true), WithSameSite(http.SameSiteStrictMode))
func (c *Cookie) Set(key string, value string, opts ...CookieOption) {
	cookie := &http.Cookie{
		Name:     key,
		Value:    value,
		MaxAge:   3600 * 24 * 365,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	for _, opt := range opts {
		opt(cookie)
	}
	http.SetCookie(c.c.Writer, cookie)
}

// SetDomain sets a cookie scoped to a specific domain.
func (c *Cookie) SetDomain(domain string, key string, value string, opts ...CookieOption) {
	c.Set(key, value, append([]CookieOption{WithDomain(domain)}, opts...)...)
}

// SetWithExpire sets a cookie with a custom max-age in seconds.
func (c *Cookie) SetWithExpire(key string, value string, expire int, opts ...CookieOption) {
	c.Set(key, value, append([]CookieOption{func(c *http.Cookie) { c.MaxAge = expire }}, opts...)...)
}

// Forever sets a session cookie (MaxAge=0, expires when browser closes).
func (c *Cookie) Forever(key string, value string, opts ...CookieOption) {
	c.Set(key, value, append([]CookieOption{func(c *http.Cookie) { c.MaxAge = 0 }}, opts...)...)
}

// ForeverDomain sets a session cookie scoped to a specific domain.
func (c *Cookie) ForeverDomain(domain string, key string, value string, opts ...CookieOption) {
	c.Set(key, value, append([]CookieOption{
		WithDomain(domain),
		func(c *http.Cookie) { c.MaxAge = 0 },
	}, opts...)...)
}

// Delete removes a cookie by setting its expiration to the past.
func (c *Cookie) Delete(key string) {
	c.Set(key, "", func(c *http.Cookie) { c.MaxAge = -1 })
}

// Expire removes a cookie by setting its expiration to the past.
func (c *Cookie) Expire(key string) {
	c.Delete(key)
}

// Update updates a cookie value with default options.
func (c *Cookie) Update(key string, value string) {
	c.Set(key, value)
}

// NewCookie creates a new Cookie helper for the given context.
func NewCookie(c *gin.Context) *Cookie {
	return &Cookie{c: c}
}

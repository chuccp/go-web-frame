// Package web: cookie helper for reading and writing HTTP cookies.
package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CookieOptions holds configurable cookie attributes.
type CookieOptions struct {
	MaxAge   int            // Max age in seconds (default: 1 year)
	Path     string         // Cookie path (default: "/")
	Domain   string         // Cookie domain (default: "")
	Secure   bool           // Secure flag: only send over HTTPS (default: false)
	HttpOnly bool           // HttpOnly flag: not accessible via JavaScript (default: true)
	SameSite http.SameSite  // SameSite attribute (default: Lax)
}

// DefaultCookieOptions returns cookie options with secure defaults.
// Secure=false for local dev; set to true in production.
func DefaultCookieOptions() CookieOptions {
	return CookieOptions{
		MaxAge:   3600 * 24 * 365,
		Path:     "/",
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
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

// Set sets a cookie with default options (HttpOnly=true, SameSite=Lax, 1 year expiry).
func (c *Cookie) Set(key string, value string) {
	c.SetWith(DefaultCookieOptions(), key, value)
}

// SetWith sets a cookie with the given options.
func (c *Cookie) SetWith(opts CookieOptions, key string, value string) {
	http.SetCookie(c.c.Writer, &http.Cookie{
		Name:     key,
		Value:    value,
		MaxAge:   opts.MaxAge,
		Path:     opts.Path,
		Domain:   opts.Domain,
		Secure:   opts.Secure,
		HttpOnly: opts.HttpOnly,
		SameSite: opts.SameSite,
	})
}

// SetDomain sets a cookie scoped to a specific domain with default options.
func (c *Cookie) SetDomain(domain string, key string, value string) {
	opts := DefaultCookieOptions()
	opts.Domain = domain
	c.SetWith(opts, key, value)
}

// SetWithExpire sets a cookie with a custom max-age in seconds.
func (c *Cookie) SetWithExpire(key string, value string, expire int) {
	opts := DefaultCookieOptions()
	opts.MaxAge = expire
	c.SetWith(opts, key, value)
}

// Forever sets a session cookie (MaxAge=0, expires when browser closes).
func (c *Cookie) Forever(key string, value string) {
	opts := DefaultCookieOptions()
	opts.MaxAge = 0
	c.SetWith(opts, key, value)
}

// ForeverDomain sets a session cookie scoped to a specific domain.
func (c *Cookie) ForeverDomain(domain string, key string, value string) {
	opts := DefaultCookieOptions()
	opts.MaxAge = 0
	opts.Domain = domain
	c.SetWith(opts, key, value)
}

// Delete removes a cookie by setting its expiration to the past.
func (c *Cookie) Delete(key string) {
	c.SetWith(CookieOptions{
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
	}, key, "")
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

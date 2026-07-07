// Package web: cookie helper for reading and writing HTTP cookies.
package web

import (
	"github.com/gin-gonic/gin"
)

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

// Set sets a cookie with a default expiration of 1 year.
func (c *Cookie) Set(key string, value string) {
	c.c.SetCookie(key, value, 3600*24*365, "/", "", false, false)
}

// SetDomain sets a cookie scoped to a specific domain with a default expiration of 1 year.
func (c *Cookie) SetDomain(domain string, key string, value string) {
	c.c.SetCookie(key, value, 3600*24*365, "/", domain, false, false)
}

// Delete removes a cookie by setting its expiration to the past.
func (c *Cookie) Delete(key string) {
	c.c.SetCookie(key, "", -1, "/", "", false, true)
}

// Update updates a cookie value with a default expiration of 1 year.
func (c *Cookie) Update(key string, value string) {
	c.c.SetCookie(key, value, 3600*24*365, "/", "", false, false)
}

// Expire removes a cookie by setting its expiration to the past.
func (c *Cookie) Expire(key string) {
	c.c.SetCookie(key, "", -1, "/", "", false, true)
}

// SetWithExpire sets a cookie with a custom max-age in seconds.
func (c *Cookie) SetWithExpire(key string, value string, expire int) {
	c.c.SetCookie(key, value, expire, "/", "", false, false)
}

// Forever sets a cookie with max-age 0 (session cookie, expires when browser closes).
func (c *Cookie) Forever(key string, value string) {
	c.c.SetCookie(key, value, 0, "/", "", false, false)
}

// ForeverDomain sets a session cookie scoped to a specific domain.
func (c *Cookie) ForeverDomain(domain string, key string, value string) {
	c.c.SetCookie(key, value, 0, "/", domain, false, false)
}

// NewCookie creates a new Cookie helper for the given context.
func NewCookie(c *gin.Context) *Cookie {
	return &Cookie{c: c}
}

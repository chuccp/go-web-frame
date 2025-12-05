package web

import (
	"github.com/gin-gonic/gin"
)

type Cookie struct {
	c *gin.Context
}

func (c *Cookie) Get(key string) string {
	cookie, err := c.c.Cookie(key)
	if err != nil {
		return ""
	}
	return cookie
}
func (c *Cookie) Set(key string, value string) {
	c.c.SetCookie(key, value, 3600*24*365, "/", "", false, true)
}
func (c *Cookie) SetDomain(domain string, key string, value string) {
	c.c.SetCookie(key, value, 3600*24*365, "/", domain, false, true)
}

func (c *Cookie) Delete(key string) {
	c.c.SetCookie(key, "", -1, "/", "", false, true)
}

func (c *Cookie) Update(key string, value string) {
	c.c.SetCookie(key, value, 3600*24*365, "/", "", false, true)
}

func (c *Cookie) Expire(key string) {
	c.c.SetCookie(key, "", -1, "/", "", false, true)
}

func (c *Cookie) SetWithExpire(key string, value string, expire int) {
	c.c.SetCookie(key, value, expire, "/", "", false, true)
}
func (c *Cookie) Forever(key string, value string) {
	c.c.SetCookie(key, value, 0, "/", "", false, true)
}
func (c *Cookie) ForeverDomain(domain string, key string, value string) {
	c.c.SetCookie(key, value, 0, "/", domain, false, true)
}
func NewCookie(c *gin.Context) *Cookie {
	return &Cookie{c: c}
}

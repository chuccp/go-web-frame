package db

import (
	"net/url"

	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/util"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

const (
	MYSQL  = "mysql"
	SQLITE = "sqlite"
)

type DB interface {
	Connection(cfg *config.Config) (db *gorm.DB, err error)
}

type noConfigDBError struct {
}

func (e *noConfigDBError) Error() string {
	return "no config db"
}

type configDBError struct {
}

func (e *configDBError) Error() string {
	return "config db error"
}

var NoConfigDBError = &noConfigDBError{}

var ConfigDBError = &configDBError{}

var dbMap = map[string]DB{
	MYSQL: &Mysql{},
}

func InitDB(c *config.Config) (*gorm.DB, error) {
	type_ := c.GetString("db.type")
	url := c.GetString("db.url")
	if util.IsNotBlank(type_) && util.IsNotBlank(url) {
		for key, db := range dbMap {
			if util.EqualsAnyIgnoreCase(type_, key) || util.StartsWithAnyIgnoreCase(url, key) {
				return db.Connection(c)
			}
		}
		return nil, ConfigDBError
	}
	return nil, NoConfigDBError
}
func getUrl(cfg *config.Config) (*url.URL, error) {
	newURL := &url.URL{}
	url_ := cfg.GetString("db.url")
	if util.IsNotBlank(url_) {
		uu, err := url.Parse(url_)
		if err != nil {
			return nil, err
		}
		newURL = uu
	}
	user := cfg.GetString("db.user")
	password := cfg.GetString("db.password")
	if util.IsNotBlank(user) || util.IsNotBlank(password) {
		if util.IsBlank(user) {
			user = newURL.User.Username()
		}
		if util.IsBlank(password) {
			password, _ = newURL.User.Password()
		}
		newURL.User = url.UserPassword(user, password)
	}

	host := cfg.GetString("db.host")
	if util.IsNotBlank(host) {
		newURL.Host = host
	}
	port := cfg.GetInt("db.port")
	if port > 0 {
		newURL.Host = newURL.Host + ":" + cast.ToString(port)
	}
	dbname := cfg.GetString("db.dbname")
	if util.IsNotBlank(dbname) {
		newURL.Path = "/" + dbname
	}
	return newURL, nil
}

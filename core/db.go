package core

import (
	"github.com/chuccp/go-web-frame/util"
	"gorm.io/gorm"
)

type noConfigDBError struct {
}

func (e *noConfigDBError) Error() string {
	return "no config db"
}

var NoConfigDBError = &noConfigDBError{}

func initDB(c *Config) (*gorm.DB, error) {
	username := c.GetString("mysql.username")
	password := c.GetString("mysql.password")
	host := c.GetString("mysql.host")
	port := c.GetInt("mysql.port")
	database := c.GetString("mysql.database")
	if util.IsBlank(username) || util.IsBlank(password) || util.IsBlank(host) || util.IsBlank(database) {
		return nil, NoConfigDBError
	}
	db, err := util.CreateMysqlConnection(username, password, host, port, database, "utf8")
	if err != nil {
		return nil, err
	}
	return db, nil
}

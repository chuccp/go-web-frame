package core

import (
	"github.com/chuccp/go-web-frame/util"
	"gorm.io/gorm"
)

func initDB(c *Config) (*gorm.DB, error) {
	username := c.GetString("mysql.username")
	password := c.GetString("mysql.password")
	host := c.GetString("mysql.host")
	port := c.GetInt("mysql.port")
	database := c.GetString("mysql.database")
	db, err := util.CreateMysqlConnection(username, password, host, port, database, "utf8")
	if err != nil {
		return nil, err
	}
	return db, nil
}

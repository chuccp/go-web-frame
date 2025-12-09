package db

import (
	"github.com/chuccp/go-web-frame/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MysqlConfigDBError struct {
}

func (e *MysqlConfigDBError) Error() string {
	return "config db error"
}

type Mysql struct {
}

func (ms *Mysql) Connection(cfg *config.Config) (db *gorm.DB, err error) {
	newURL, err := getUrl(cfg)
	if err != nil {
		return nil, err
	}
	newURL.Scheme = "mysql"
	dsn := newURL.String()
	return gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
}

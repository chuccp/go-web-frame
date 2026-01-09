package db

import (
	"fmt"

	"emperror.dev/errors"
	log2 "github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MysqlConfig struct {
	Dbname   string
	Database string
	Charset  string
	Username string
	User     string
	Password string
	Host     string
	Port     int
}

func (mysqlConfig *MysqlConfig) Connection() (db *DB, err error) {
	if util.IsBlank(mysqlConfig.Username) {
		mysqlConfig.Username = mysqlConfig.User
	}
	if util.IsBlank(mysqlConfig.Database) {
		mysqlConfig.Database = mysqlConfig.Dbname
	}
	if mysqlConfig.Port == 0 {
		mysqlConfig.Port = 3306
	}
	if util.IsBlank(mysqlConfig.Charset) {
		mysqlConfig.Charset = "utf8"
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", mysqlConfig.Username, mysqlConfig.Password, mysqlConfig.Host, mysqlConfig.Port, mysqlConfig.Database, mysqlConfig.Charset)
	log2.Debug("mysql", zap.String("dsn", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", mysqlConfig.Username, "******", mysqlConfig.Host, mysqlConfig.Port, mysqlConfig.Database, mysqlConfig.Charset)))
	db_, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	return &DB{db: db_}, err
}

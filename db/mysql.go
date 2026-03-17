package db

import (
	"fmt"
	"time"

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
	// Connection pool settings
	MaxOpenConns    int `mapstructure:"max_open_conns"`
	MaxIdleConns    int `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int `mapstructure:"conn_max_lifetime"` // in seconds
}

func ConnectionMysql(host string, port int, username string, password string, dbname string, charset string) (db *DB, err error) {
	var mysqlConfig = &MysqlConfig{Host: host, Port: port, Username: username, Password: password, Dbname: dbname, Charset: charset}
	return mysqlConfig.Connection()

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
	// Set default connection pool values if not specified
	if mysqlConfig.MaxOpenConns == 0 {
		mysqlConfig.MaxOpenConns = 100
	}
	if mysqlConfig.MaxIdleConns == 0 {
		mysqlConfig.MaxIdleConns = 10
	}
	if mysqlConfig.ConnMaxLifetime == 0 {
		mysqlConfig.ConnMaxLifetime = 3600 // 1 hour
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", mysqlConfig.Username, mysqlConfig.Password, mysqlConfig.Host, mysqlConfig.Port, mysqlConfig.Database, mysqlConfig.Charset)
	log2.Debug("mysql", zap.String("dsn", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", mysqlConfig.Username, "******", mysqlConfig.Host, mysqlConfig.Port, mysqlConfig.Database, mysqlConfig.Charset)))
	db_, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	// Configure connection pool
	sqlDB, err := db_.DB()
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	if mysqlConfig.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(mysqlConfig.MaxOpenConns)
	}
	if mysqlConfig.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(mysqlConfig.MaxIdleConns)
	}
	if mysqlConfig.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(mysqlConfig.ConnMaxLifetime) * time.Second)
	}
	return &DB{db: db_}, err
}

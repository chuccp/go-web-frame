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

// MysqlConfig holds MySQL connection configuration.
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

// ConnectionMysql creates a new MySQL database connection.
func ConnectionMysql(host string, port int, username string, password string, dbname string, charset string) (db *DB, err error) {
	var mysqlConfig = &MysqlConfig{Host: host, Port: port, Username: username, Password: password, Dbname: dbname, Charset: charset}
	return mysqlConfig.Connection()

}

func (mysqlConfig *MysqlConfig) GetMaxOpenConns() int {
	if mysqlConfig.MaxOpenConns == 0 {
		return 100
	}
	return mysqlConfig.MaxOpenConns
}

func (mysqlConfig *MysqlConfig) GetMaxIdleConns() int {
	if mysqlConfig.MaxIdleConns == 0 {
		return 10
	}
	return mysqlConfig.MaxIdleConns
}

func (mysqlConfig *MysqlConfig) GetConnMaxLifetime() int {
	if mysqlConfig.ConnMaxLifetime == 0 {
		return 3600 // 1 hour
	}
	return mysqlConfig.ConnMaxLifetime
}

// Connection creates a MySQL database connection from this config.
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
		mysqlConfig.Charset = "utf8mb4"
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", mysqlConfig.Username, mysqlConfig.Password, mysqlConfig.Host, mysqlConfig.Port, mysqlConfig.Database, mysqlConfig.Charset)
	log2.Debug("mysql", zap.String("dsn", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", mysqlConfig.Username, "******", mysqlConfig.Host, mysqlConfig.Port, mysqlConfig.Database, mysqlConfig.Charset)))
	db_, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	if err := db_.Use(&ZeroTimePlugin{}); err != nil {
		return nil, errors.WithStackIf(err)
	}
	if err := ApplyConnectionPool(db_, mysqlConfig); err != nil {
		return nil, err
	}
	return &DB{db: db_}, err
}

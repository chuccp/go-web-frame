package db

import (
	"time"

	"emperror.dev/errors"
	log2 "github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SQLiteConfig struct {
	FilePath string
	// Connection pool settings
	MaxOpenConns    int `mapstructure:"max_open_conns"`
	MaxIdleConns    int `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int `mapstructure:"conn_max_lifetime"` // in seconds
}

func ConnectionSQLite(FilePath string) (db *DB, err error) {
	return (&SQLiteConfig{FilePath: FilePath}).Connection()
}

func (sqliteConfig *SQLiteConfig) Connection() (db *DB, err error) {
	// Set default connection pool values if not specified
	if sqliteConfig.MaxOpenConns == 0 {
		sqliteConfig.MaxOpenConns = 10
	}
	if sqliteConfig.MaxIdleConns == 0 {
		sqliteConfig.MaxIdleConns = 5
	}
	if sqliteConfig.ConnMaxLifetime == 0 {
		sqliteConfig.ConnMaxLifetime = 3600 // 1 hour
	}
	log2.Debug("sqlite", zap.String("dsn", sqliteConfig.FilePath))
	sb, err := gorm.Open(sqlite.Open(sqliteConfig.FilePath), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	// Configure connection pool
	sqlDB, err := sb.DB()
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	if sqliteConfig.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(sqliteConfig.MaxOpenConns)
	}
	if sqliteConfig.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(sqliteConfig.MaxIdleConns)
	}
	if sqliteConfig.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(sqliteConfig.ConnMaxLifetime) * time.Second)
	}
	return &DB{db: sb}, err
}

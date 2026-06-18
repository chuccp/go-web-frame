package db

import (
	"emperror.dev/errors"
	log2 "github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SQLiteConfig holds SQLite connection configuration.
type SQLiteConfig struct {
	FilePath string `mapstructure:"path"`
	// Connection pool settings
	MaxOpenConns    int `mapstructure:"max_open_conns"`
	MaxIdleConns    int `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int `mapstructure:"conn_max_lifetime"` // in seconds
}

// ConnectionSQLite creates a new SQLite database connection.
func ConnectionSQLite(FilePath string) (db *DB, err error) {
	return (&SQLiteConfig{FilePath: FilePath}).Connection()
}

func (sqliteConfig *SQLiteConfig) GetMaxOpenConns() int {
	if sqliteConfig.MaxOpenConns == 0 {
		return 10
	}
	return sqliteConfig.MaxOpenConns
}

func (sqliteConfig *SQLiteConfig) GetMaxIdleConns() int {
	if sqliteConfig.MaxIdleConns == 0 {
		return 5
	}
	return sqliteConfig.MaxIdleConns
}

func (sqliteConfig *SQLiteConfig) GetConnMaxLifetime() int {
	if sqliteConfig.ConnMaxLifetime == 0 {
		return 3600 // 1 hour
	}
	return sqliteConfig.ConnMaxLifetime
}

// Connection creates a SQLite database connection from this config.
func (sqliteConfig *SQLiteConfig) Connection() (db *DB, err error) {
	log2.Debug("sqlite", zap.String("dsn", sqliteConfig.FilePath))
	sb, err := gorm.Open(sqlite.Open(sqliteConfig.FilePath), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	if err := ApplyConnectionPool(sb, sqliteConfig); err != nil {
		return nil, err
	}
	return &DB{db: sb}, err
}

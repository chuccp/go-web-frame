package db

import (
	"emperror.dev/errors"
	log2 "github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SQLiteConfig struct {
	FilePath string
}

func (sqliteConfig *SQLiteConfig) Connection() (db *DB, err error) {
	log2.Debug("sqlite", zap.String("dsn", sqliteConfig.FilePath))
	sb, err := gorm.Open(sqlite.Open(sqliteConfig.FilePath), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	return &DB{db: sb}, err
}

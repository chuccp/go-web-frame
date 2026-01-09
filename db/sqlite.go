package db

import (
	log2 "github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SQLiteConfig struct {
	FilePath string
}
type SQLite struct {
}

func (sq *SQLite) Connection(sqliteConfig *SQLiteConfig) (db *gorm.DB, err error) {
	//sqliteConfig := &SQLiteConfig{}
	//err = cfg.Unmarshal("web.db", sqliteConfig)
	//if err != nil {
	//	return nil, err
	//}
	log2.Debug("sqlite", zap.String("dsn", sqliteConfig.FilePath))
	return gorm.Open(sqlite.Open(sqliteConfig.FilePath), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
}

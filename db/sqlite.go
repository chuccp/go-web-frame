package db

import (
	"github.com/chuccp/go-web-frame/config"
	"gorm.io/gorm"
)

type SQLite struct {
}

func (ms *SQLite) Connection(cfg *config.Config) (db *gorm.DB, err error) {
	return nil, nil
}

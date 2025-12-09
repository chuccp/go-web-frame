package db

import (
	"github.com/chuccp/go-web-frame/config"
	log2 "github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	MYSQL  = "mysql"
	SQLITE = "sqlite"
)

type DB interface {
	Connection(cfg *config.Config, log *log2.Logger) (db *gorm.DB, err error)
}

type noConfigDBError struct {
}

func (e *noConfigDBError) Error() string {
	return "no config db"
}

type configDBError struct {
}

func (e *configDBError) Error() string {
	return "config db error"
}

var NoConfigDBError = &noConfigDBError{}

var ConfigDBError = &configDBError{}

var dbMap = map[string]DB{
	MYSQL: &Mysql{},
}

func InitDB(c *config.Config, log *log2.Logger) (*gorm.DB, error) {
	type_ := c.GetString("web.db.type")
	log.Info("db type", zap.String("type", type_))
	if util.IsNotBlank(type_) {
		for key, db := range dbMap {
			if util.EqualsAnyIgnoreCase(type_, key) {
				return db.Connection(c, log)
			}
		}
		return nil, ConfigDBError
	}
	return nil, NoConfigDBError
}

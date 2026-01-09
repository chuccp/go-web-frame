package db

import (
	"emperror.dev/errors"
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

type Source interface {
	Connection(cfg config.IConfig) (db *gorm.DB, err error)
}
type Session struct {
	db *DB
}

func (s *Session) Delete(value any, conds ...any) error {
	err := s.db.Delete(value, conds...)
	return err
}

type DB struct {
	db *gorm.DB
}

func (d *DB) Transaction(fc func(tx *DB) error) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		return fc(&DB{db: tx})
	})
}

func (d *DB) Migrator() gorm.Migrator {
	return d.db.Migrator()
}

func (d *DB) Table(name string) *DB {
	tx := d.db.Table(name)
	return &DB{db: tx}
}

func (d *DB) Session(g *gorm.Session) *Session {
	return &Session{db: &DB{db: d.db.Session(g)}}
}

func (d *DB) AutoMigrate(t ...any) error {
	return d.db.AutoMigrate(t)
}

func (d *DB) Delete(value any, conds ...any) error {
	tx := d.db.Delete(value, conds...)
	return tx.Error
}

func (d *DB) Save(entry any) error {
	tx := d.db.Save(entry)
	return tx.Error
}

func (d *DB) Create(value any) error {
	tx := d.db.Create(value)
	return tx.Error
}

func (d *DB) Where(query any, args ...any) *DB {
	return &DB{db: d.db.Where(query, args...)}
}

func (d *DB) Offset(i int) *DB {
	tx := d.db.Offset(i)
	return &DB{db: tx}
}

func (d *DB) Order(query any) *DB {
	tx := d.db.Order(query)
	return &DB{db: tx}
}

func (d *DB) Limit(size int) *DB {
	tx := d.db.Limit(size)
	return &DB{db: tx}
}

func (d *DB) Find(dest any, conds ...any) error {
	tx := d.db.Find(dest, conds...)
	return tx.Error

}

func (d *DB) First(dest any, conds ...any) error {
	tx := d.db.First(dest, conds...)
	return tx.Error
}

func (d *DB) Count(i *int64) error {
	tx := d.db.Count(i)
	return tx.Error
}

func (d *DB) Updates(values any) error {
	tx := d.db.Updates(values)
	return tx.Error
}

func (d *DB) UpdateColumn(column string, value any) error {
	tx := d.db.UpdateColumn(column, value)
	return tx.Error
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

var dbMap = map[string]Source{
	MYSQL:  &Mysql{},
	SQLITE: &SQLite{},
}

func InitDB(c config.IConfig) (*DB, error) {
	type_ := c.GetString("web.db.type")
	log2.Info("db type", zap.String("type", type_))
	if util.IsNotBlank(type_) {
		for key, db := range dbMap {
			if util.EqualsAnyIgnoreCase(type_, key) {
				connection, err := db.Connection(c)
				if err != nil {
					return nil, errors.WithStackIf(err)
				}
				return &DB{db: connection}, nil
			}
		}
		return nil, errors.WithStackIf(ConfigDBError)
	}
	return nil, errors.WithStackIf(NoConfigDBError)
}

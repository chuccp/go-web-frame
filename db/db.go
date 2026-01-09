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
	db *gorm.DB
}

func (s *Session) Delete(value any, conds ...any) error {
	tx := s.db.Delete(value, conds...)
	return tx.Error
}

type Table struct {
	db *gorm.DB
}

func (t *Table) Session(g *gorm.Session) *Session {
	return &Session{db: t.db.Session(g)}
}

func (t *Table) AutoMigrate(v ...any) error {
	return t.db.AutoMigrate(v)
}

func (t *Table) Delete(value any, conds ...any) error {
	tx := t.db.Delete(value, conds...)
	return tx.Error
}

func (t *Table) Save(entry any) error {
	tx := t.db.Save(entry)
	return tx.Error
}

func (t *Table) Create(value any) error {
	tx := t.db.Create(value)
	return tx.Error
}

func (t *Table) Where(query any, args ...any) *Table {
	t.db = t.db.Where(query, args...)
	return t
}

func (t *Table) Offset(i int) *Table {
	tx := t.db.Offset(i)
	return &Table{db: tx}
}

func (t *Table) Order(query any) *Table {
	tx := t.db.Order(query)
	return &Table{db: tx}
}

func (t *Table) Limit(size int) *Table {
	tx := t.db.Limit(size)
	return &Table{db: tx}
}

func (t *Table) Find(dest any, conds ...any) error {
	tx := t.db.Find(dest, conds...)
	return tx.Error

}

func (t *Table) First(dest any, conds ...any) error {
	tx := t.db.First(dest, conds...)
	return tx.Error
}

func (t *Table) Count(i *int64) error {
	tx := t.db.Count(i)
	return tx.Error
}

func (t *Table) Updates(values any) error {
	tx := t.db.Updates(values)
	return tx.Error
}

func (t *Table) UpdateColumn(column string, value any) error {
	tx := t.db.UpdateColumn(column, value)
	return tx.Error
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

func (d *DB) Table(name string) *Table {
	tx := d.db.Table(name)
	return &Table{db: tx}
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

package db

import (
	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/util"
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

type IConfig interface {
	Connection() (*DB, error)
}
type Config struct {
	Type string
}

const ConfigKey = "web.db"

func CreateDB(c config.IConfig) (*DB, error) {
	var config2 Config
	err := c.Unmarshal(ConfigKey, &config2)
	if err != nil {
		return nil, err
	}
	if util.IsNotBlank(config2.Type) {
		if util.EqualsAnyIgnoreCase(config2.Type, MYSQL) {
			var mysqlConfig MysqlConfig
			err := c.Unmarshal(ConfigKey, &mysqlConfig)
			if err != nil {
				return nil, err
			}
			return mysqlConfig.Connection()
		}
		if util.EqualsAnyIgnoreCase(config2.Type, SQLITE) {
			var sqliteConfig SQLiteConfig
			err := c.Unmarshal(ConfigKey, &sqliteConfig)
			if err != nil {
				return nil, err
			}
			return sqliteConfig.Connection()
		}
	}
	return nil, errors.WithStackIf(NoConfigDBError)
}

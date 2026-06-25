package db

import (
	"reflect"
	"time"

	"gorm.io/gorm"
)

// ZeroTimePlugin is a GORM plugin that automatically converts zero time.Time values
// to NULL before Create and Update operations. This prevents MySQL strict mode from
// rejecting '0000-00-00' datetime values.
type ZeroTimePlugin struct{}

// Name returns the plugin name for GORM registration.
func (p *ZeroTimePlugin) Name() string {
	return "zerotime"
}

// Initialize registers the BeforeCreate and BeforeUpdate callbacks.
func (p *ZeroTimePlugin) Initialize(db *gorm.DB) error {
	db.Callback().Create().Before("gorm:create").Register("zerotime:before_create", beforeCreateCallback)
	db.Callback().Update().Before("gorm:update").Register("zerotime:before_update", beforeUpdateCallback)
	return nil
}

// beforeCreateCallback processes zero time values before Create operations.
func beforeCreateCallback(db *gorm.DB) {
	processZeroTime(db)
}

// beforeUpdateCallback processes zero time values before Update operations.
func beforeUpdateCallback(db *gorm.DB) {
	processZeroTime(db)
}

// processZeroTime scans the statement's Dest for time.Time fields and sets zero values to nil.
func processZeroTime(db *gorm.DB) {
	if db.Statement.Schema == nil {
		return
	}

	dest := db.Statement.Dest
	if dest == nil {
		return
	}

	// Handle pointer to struct
	val := reflect.ValueOf(dest)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return
		}
		val = val.Elem()
	}

	// Only process struct types
	if val.Kind() != reflect.Struct {
		return
	}

	// Use GORM's schema to iterate over fields
	for _, field := range db.Statement.Schema.Fields {
		if field.FieldType != reflect.TypeOf(time.Time{}) {
			continue
		}

		fieldVal := val.FieldByName(field.Name)
		if !fieldVal.IsValid() || !fieldVal.CanSet() {
			continue
		}

		t, ok := fieldVal.Interface().(time.Time)
		if !ok {
			continue
		}

		// If zero time, set to current time
		if t.IsZero() {
			db.Statement.SetColumn(field.Name, time.Now())
		}
	}
}

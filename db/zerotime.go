package db

import (
	"reflect"
	"time"

	"gorm.io/gorm"
)

// ZeroTimePlugin is a GORM plugin that automatically converts zero time.Time values
// to the current time before Create and Update operations. This prevents MySQL strict
// mode from rejecting '0000-00-00' datetime values.
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

// processZeroTime scans the statement's Dest for time.Time fields and sets zero values to current time.
func processZeroTime(db *gorm.DB) {
	dest := db.Statement.Dest
	if dest == nil {
		return
	}

	val := reflect.ValueOf(dest)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return
		}
		val = val.Elem()
	}

	// Handle map[string]interface{} (used in UpdateForMap)
	if val.Kind() == reflect.Map {
		processMapZeroTime(db, val)
		return
	}

	// Handle struct (used in Save/Create/Update with struct)
	if val.Kind() == reflect.Struct && db.Statement.Schema != nil {
		processStructZeroTime(db, val)
	}
}

// processMapZeroTime handles zero time values in map[string]interface{} destinations.
func processMapZeroTime(db *gorm.DB, val reflect.Value) {
	now := time.Now()

	// Get the original map
	origMap, ok := db.Statement.Dest.(map[string]interface{})
	if !ok {
		return
	}

	// Check and update zero time values
	hasChanges := false
	for key, value := range origMap {
		t, isTime := value.(time.Time)
		if isTime && t.IsZero() {
			origMap[key] = now
			hasChanges = true
		}
	}

	// If we made changes, reassign the map to the statement
	if hasChanges {
		db.Statement.Dest = origMap
	}
}

// processStructZeroTime handles zero time values in struct destinations using GORM schema.
func processStructZeroTime(db *gorm.DB, val reflect.Value) {
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

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFromBytes_JSON(t *testing.T) {
	// Given: JSON data
	data := []byte(`{
		"web": {
			"db": {
				"type": "mysql",
				"host": "localhost",
				"port": 3306,
				"max_open_conns": 100
			}
		},
		"server": {
			"port": 8080
		}
	}`)

	// When: creating config from bytes
	cfg, err := NewFromBytes(data, "json")

	// Then: should success and values correct
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "mysql", cfg.GetString("web.db.type"))
	assert.Equal(t, "localhost", cfg.GetString("web.db.host"))
	assert.Equal(t, 3306, cfg.GetInt("web.db.port"))
	assert.Equal(t, 100, cfg.GetInt("web.db.max_open_conns"))
	assert.Equal(t, 8080, cfg.GetInt("server.port"))
}

func TestNewFromBytes_YAML(t *testing.T) {
	// Given: YAML data
	data := []byte(`
web:
  db:
    type: sqlite
    file_path: ":memory:"
    max_open_conns: 10
server:
  port: 8080
`)

	// When: creating config from bytes
	cfg, err := NewFromBytes(data, "yaml")

	// Then: should success
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "sqlite", cfg.GetString("web.db.type"))
	assert.Equal(t, 10, cfg.GetInt("web.db.max_open_conns"))
	assert.Equal(t, 8080, cfg.GetInt("server.port"))
}

func TestNewFromBytes_InvalidFormat(t *testing.T) {
	// Given: invalid JSON
	data := []byte(`{ invalid json`)

	// When: creating config
	cfg, err := NewFromBytes(data, "json")

	// Then: should error
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestNewFromBytes_GetStringOrDefault(t *testing.T) {
	data := []byte(`{"key": "value"}`)
	cfg, _ := NewFromBytes(data, "json")

	// Existing key
	assert.Equal(t, "value", cfg.GetStringOrDefault("key", "default"))
	// Non-existent key
	assert.Equal(t, "default", cfg.GetStringOrDefault("missing", "default"))
}

func TestNewFromBytes_GetIntOrDefault(t *testing.T) {
	data := []byte(`{"port": 8080}`)
	cfg, _ := NewFromBytes(data, "json")

	// Existing key
	assert.Equal(t, 8080, cfg.GetIntOrDefault("port", 80))
	// Non-existent key
	assert.Equal(t, 80, cfg.GetIntOrDefault("missing", 80))
}

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()
	assert.NotNil(t, cfg)
}

func TestConfig_Put(t *testing.T) {
	cfg := NewConfig()
	cfg.Put("test.key", "value")
	assert.Equal(t, "value", cfg.GetString("test.key"))
}

func TestConfig_HasKey(t *testing.T) {
	cfg := NewConfig()
	cfg.Put("test.key", "value")
	assert.True(t, cfg.HasKey("test.key"))
	assert.False(t, cfg.HasKey("test.missing"))
}

func TestMergeConfig_Single(t *testing.T) {
	cfg := NewConfig()
	cfg.Put("key", "value")
	cfg.Put("nested.key", 42)

	merged := MergeConfig(cfg)
	assert.NotNil(t, merged)
	assert.Equal(t, "value", merged.GetString("key"))
	assert.Equal(t, 42, merged.GetInt("nested.key"))
}

func TestMergeConfig_Merge(t *testing.T) {
	cfg1 := NewConfig()
	cfg1.Put("shared", "from_first")
	cfg1.Put("only_in_first", "first")

	cfg2 := NewConfig()
	cfg2.Put("shared", "from_second")
	cfg2.Put("only_in_second", "second")

	merged := MergeConfig(cfg1, cfg2)

	// Later config overrides
	assert.Equal(t, "from_second", merged.GetString("shared"))
	// Non-overlapping keys preserved
	assert.Equal(t, "first", merged.GetString("only_in_first"))
	assert.Equal(t, "second", merged.GetString("only_in_second"))
}

func TestMergeConfig_Empty(t *testing.T) {
	merged := MergeConfig()
	assert.NotNil(t, merged)
	assert.False(t, merged.HasKey("any"))
}

func TestMergeConfig_OverridesInOrder(t *testing.T) {
	cfg1 := NewConfig()
	cfg1.Put("a", 1)
	cfg1.Put("b", 2)

	cfg2 := NewConfig()
	cfg2.Put("b", 20)
	cfg2.Put("c", 30)

	cfg3 := NewConfig()
	cfg3.Put("c", 300)
	cfg3.Put("d", 400)

	merged := MergeConfig(cfg1, cfg2, cfg3)
	assert.Equal(t, 1, merged.GetInt("a"))   // from cfg1
	assert.Equal(t, 20, merged.GetInt("b"))  // cfg2 overrides cfg1
	assert.Equal(t, 300, merged.GetInt("c")) // cfg3 overrides cfg2
	assert.Equal(t, 400, merged.GetInt("d")) // from cfg3
}

func TestMergeConfig_NestedSettings(t *testing.T) {
	cfg1 := NewConfig()
	cfg1.Put("db.type", "mysql")
	cfg1.Put("db.host", "localhost")

	cfg2 := NewConfig()
	cfg2.Put("db.host", "production-host")
	cfg2.Put("server.port", 8080)

	merged := MergeConfig(cfg1, cfg2)
	assert.Equal(t, "mysql", merged.GetString("db.type"))
	assert.Equal(t, "production-host", merged.GetString("db.host"))
	assert.Equal(t, 8080, merged.GetInt("server.port"))
}

func TestMergeConfig_DoesNotMutateOriginals(t *testing.T) {
	cfg1 := NewConfig()
	cfg1.Put("key", "original")

	cfg2 := NewConfig()
	cfg2.Put("key", "override")

	MergeConfig(cfg1, cfg2)

	// Originals unchanged
	assert.Equal(t, "original", cfg1.GetString("key"))
	assert.Equal(t, "override", cfg2.GetString("key"))
}

func TestConfig_UnmarshalKey(t *testing.T) {
	data := []byte(`{"db": {"type": "mysql", "port": 3306}}`)
	cfg, _ := NewFromBytes(data, "json")

	type DBConfig struct {
		Type string
		Port int
	}
	var dbConfig DBConfig
	err := cfg.UnmarshalKey("db", &dbConfig)

	assert.NoError(t, err)
	assert.Equal(t, "mysql", dbConfig.Type)
	assert.Equal(t, 3306, dbConfig.Port)
}

func TestSingleFileConfig_WriteConfig(t *testing.T) {
	// Create a temp config file
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	initial := []byte(`{"server":{"port":8080},"db":{"type":"sqlite"}}`)
	err := os.WriteFile(path, initial, 0644)
	assert.NoError(t, err)

	// Load the config
	sfc, err := LoadSingleFileConfig(path)
	assert.NoError(t, err)
	assert.NotNil(t, sfc)

	// Verify initial values
	assert.Equal(t, 8080, sfc.GetInt("server.port"))
	assert.Equal(t, "sqlite", sfc.GetString("db.type"))

	// Modify config
	sfc.Put("server.port", 9090)
	sfc.Put("db.host", "localhost")

	// Write back
	err = sfc.WriteConfig()
	assert.NoError(t, err)

	// Reload and verify changes persisted
	sfc2, err := LoadSingleFileConfig(path)
	assert.NoError(t, err)
	assert.Equal(t, 9090, sfc2.GetInt("server.port"))
	assert.Equal(t, "sqlite", sfc2.GetString("db.type"))
	assert.Equal(t, "localhost", sfc2.GetString("db.host"))
}

func TestSingleFileConfig_WriteConfig_JSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	initial := []byte(`{"app":{"name":"test"}}`)
	err := os.WriteFile(path, initial, 0644)
	assert.NoError(t, err)

	sfc, err := LoadSingleFileConfig(path)
	assert.NoError(t, err)

	sfc.Put("app.version", "1.0.0")
	err = sfc.WriteConfig()
	assert.NoError(t, err)

	// Reload and verify
	sfc2, err := LoadSingleFileConfig(path)
	assert.NoError(t, err)
	assert.Equal(t, "test", sfc2.GetString("app.name"))
	assert.Equal(t, "1.0.0", sfc2.GetString("app.version"))
}

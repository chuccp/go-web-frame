package config

import (
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

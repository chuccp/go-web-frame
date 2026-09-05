package config

import (
	"bytes"
	"path/filepath"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/chuccp/go-web-frame/value"
	"github.com/go-viper/encoding/ini"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// IConfig defines the interface for reading application configuration.
type IConfig interface {
	GetString(key string) string
	Put(key string, value any)
	HasKey(key string) bool
	GetStringOrDefault(key string, defaultValue string) string
	GetInt(key string) int
	GetIntOrDefault(key string, defaultValue int) int
	GetBoolOrDefault(key string, defaultValue bool) bool
	Unmarshal(v any) error
	UnmarshalKey(key string, v any) error
	ReplaceKey(key string, newKey string)
	WriteConfig() error
	AllSettings() *value.Object
}

// Config wraps a Viper configuration instance and implements IConfig.
type Config struct {
	//object *viper.Viper
	object *value.Object
}

// GetString returns the string value for the given key. Supports dot-separated paths.
func (c *Config) GetString(key string) string {
	v := c.object.GetByPath(key)
	if v == nil {
		return ""
	}
	if v.IsNull() {
		return ""
	}
	return v.String()
}

// Put sets a configuration value. Supports dot-separated key paths (e.g. "web.db.type").
func (c *Config) Put(key string, val any) {
	c.object.PutByPath(key, val)
}

// GetStringOrDefault returns the string value for the given key, or defaultValue if blank.
func (c *Config) GetStringOrDefault(key string, defaultValue string) string {
	s := c.GetString(key)
	if util.IsBlank(s) {
		return defaultValue
	}
	return s
}

// HasKey reports whether the given key exists in the configuration. Supports dot-separated paths.
func (c *Config) HasKey(key string) bool {
	return c.object.GetByPath(key) != nil
}

// UnmarshalKey unmarshals configuration under the given key into the target struct.
// Supports both camelCase and snake_case keys in config files.
func (c *Config) UnmarshalKey(key string, v any) error {
	val := c.object.GetByPath(key)
	if val == nil {
		return nil
	}
	return errors.WithStackIf(val.Unmarshal(v))
}

// Unmarshal unmarshals the entire configuration into the target struct.
// Supports both camelCase and snake_case keys in config files.
func (c *Config) Unmarshal(v any) error {
	return errors.WithStackIf(c.object.Unmarshal(v))
}

// GetInt returns the int value for the given key. Supports dot-separated paths.
func (c *Config) GetInt(key string) int {
	v := c.object.GetByPath(key)
	if v == nil || !v.IsNumber() {
		return 0
	}
	return int(v.AsNumber().Int64())
}

// GetIntOrDefault returns the int value for the given key, or defaultValue if not set.
func (c *Config) GetIntOrDefault(key string, defaultValue int) int {
	v := c.object.GetByPath(key)
	if v == nil || !v.IsNumber() {
		return defaultValue
	}
	return int(v.AsNumber().Int64())
}

// GetBoolOrDefault returns the bool value for the given key, or defaultValue if not set.
func (c *Config) GetBoolOrDefault(key string, defaultValue bool) bool {
	v := c.object.GetByPath(key)
	if v == nil || !v.IsBool() {
		return defaultValue
	}
	return v.String() == "true"
}

// ReplaceKey copies the value from key to newKey if key is set. Supports dot-separated paths.
func (c *Config) ReplaceKey(key string, newKey string) {
	v := c.object.GetByPath(key)
	if v != nil {
		c.object.PutByPath(newKey, v)
	}
}

// AllSettings returns the underlying configuration object.
func (c *Config) AllSettings() *value.Object {
	return c.object
}

// WriteConfig is a no-op for Config; use SingleFileConfig for file-backed writes.
func (c *Config) WriteConfig() error {
	// Config doesn't have a file to write to, this is a no-op
	return errors.Errorf("Config doesn't have a file to write to")
}

// SingleFileConfig extends Config with a file path for write-back support.
type SingleFileConfig struct {
	*Config
	path string
	v    *viper.Viper
}

// WriteConfig writes the configuration back to the file.
func (c *SingleFileConfig) WriteConfig() error {
	return c.v.WriteConfig()
}

// LoadSingleFileConfig loads a single configuration file (YAML, JSON, TOML, INI).
// Creates the file if it does not exist.
func LoadSingleFileConfig(path string) (*SingleFileConfig, error) {
	registry := viper.NewCodecRegistry()
	er := registry.RegisterCodec("ini", ini.Codec{})
	if er != nil {
		return nil, errors.WithStackIf(er)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	log.Info("Load the configuration file", zap.String("path", absPath))
	err = util.CreateFileIfNoExists(absPath)
	if err != nil {
		return nil, err
	}
	_viper_ := viper.NewWithOptions(viper.WithCodecRegistry(registry))
	_viper_.SetConfigFile(absPath)
	err = _viper_.ReadInConfig()
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	return &SingleFileConfig{Config: NewConfig(), v: _viper_, path: absPath}, nil
}

// NewConfig creates a new empty Config with a fresh Viper instance.
func NewConfig() *Config {
	return &Config{object: value.NewObject()}
}

// LoadConfig loads and merges multiple configuration files.
func LoadConfig(paths ...string) (*Config, error) {
	obj := value.NewObject()
	registry := viper.NewCodecRegistry()
	err := registry.RegisterCodec("ini", ini.Codec{})
	if err != nil {
		return nil, errors.WithStackIf(err)
	}

	for _, path := range paths {
		path, err := filepath.Abs(path)
		if err != nil {
			return nil, errors.WithStackIf(err)
		}
		viper2 := viper.NewWithOptions(viper.WithCodecRegistry(registry))
		viper2.SetConfigFile(path)
		err = viper2.ReadInConfig()
		if err != nil {
			return nil, errors.WithStackIf(err)
		}
		obj.PutMap(viper2.AllSettings())
	}
	return &Config{object: obj}, nil
}

// LoadAutoConfig creates a new empty Config, typically used when configuration
// is provided programmatically rather than from files.
func LoadAutoConfig() *Config {
	return NewConfig()
}

// MergeConfig merges multiple IConfig instances into a single *Config.
// Later configs take precedence over earlier ones for duplicate keys.
func MergeConfig(configs ...IConfig) *Config {
	obj := value.NewObject()
	for _, cfg := range configs {
		obj.AddAll(cfg.AllSettings())
	}
	return &Config{object: obj}
}

// NewFromBytes creates a Config from raw bytes with the specified format (json, yaml, toml, etc.)
func NewFromBytes(data []byte, format string) (*Config, error) {
	obj := value.NewObject()
	registry := viper.NewCodecRegistry()
	err := registry.RegisterCodec("ini", ini.Codec{})
	if err != nil {
		return nil, errors.WithStackIf(err)
	}

	v := viper.NewWithOptions(viper.WithCodecRegistry(registry))
	v.SetConfigType(format)

	err = v.ReadConfig(bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	obj.PutMap(v.AllSettings())
	return &Config{object: obj}, nil
}

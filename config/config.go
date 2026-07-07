package config

import (
	"bytes"
	"path/filepath"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
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
}

// Config wraps a Viper configuration instance and implements IConfig.
type Config struct {
	v *viper.Viper
}

// GetString returns the string value for the given key.
func (c *Config) GetString(key string) string {
	return c.v.GetString(key)
}
// Put sets a configuration value.
func (c *Config) Put(key string, value any) {
	c.v.Set(key, value)
}
// GetStringOrDefault returns the string value for the given key, or defaultValue if blank.
func (c *Config) GetStringOrDefault(key string, defaultValue string) string {
	v := c.v.GetString(key)
	if util.IsBlank(v) {
		return defaultValue
	}
	return v
}
// HasKey reports whether the given key exists in the configuration.
func (c *Config) HasKey(key string) bool {
	return c.v.IsSet(key)
}
// UnmarshalKey unmarshals configuration under the given key into the target struct.
func (c *Config) UnmarshalKey(key string, v any) error {
	return errors.WithStackIf(c.v.UnmarshalKey(key, v))
}

// Unmarshal unmarshals the entire configuration into the target struct.
func (c *Config) Unmarshal(v any) error {
	return errors.WithStackIf(c.v.Unmarshal(v))
}

// GetInt returns the int value for the given key.
func (c *Config) GetInt(key string) int {
	return c.v.GetInt(key)
}

// GetIntOrDefault returns the int value for the given key, or defaultValue if not set.
func (c *Config) GetIntOrDefault(key string, defaultValue int) int {
	if !c.v.IsSet(key) {
		return defaultValue
	}
	return c.v.GetInt(key)
}
// GetBoolOrDefault returns the bool value for the given key, or defaultValue if not set.
func (c *Config) GetBoolOrDefault(key string, defaultValue bool) bool {
	if util.IsBlank(key) || !c.v.IsSet(key) {
		return defaultValue
	}
	return c.v.GetBool(key)
}
// ReplaceKey copies the value from key to newKey if key is set.
func (c *Config) ReplaceKey(key string, newKey string) {
	if c.v.IsSet(key) {
		c.v.Set(newKey, c.v.Get(key))
	}
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
		return nil, errors.WithStackIf(er)
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
	return &SingleFileConfig{Config: &Config{v: _viper_}, path: absPath}, nil
}

// NewConfig creates a new empty Config with a fresh Viper instance.
func NewConfig() *Config {
	return &Config{v: viper.New()}
}
// LoadConfig loads and merges multiple configuration files.
func LoadConfig(paths ...string) (*Config, error) {
	registry := viper.NewCodecRegistry()
	err := registry.RegisterCodec("ini", ini.Codec{})
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	_viper_ := viper.New()
	for _, path := range paths {
		viper2 := viper.NewWithOptions(viper.WithCodecRegistry(registry))
		viper2.SetConfigFile(path)
		err := viper2.ReadInConfig()
		if err != nil {
			return nil, errors.WithStackIf(err)
		}
		err = _viper_.MergeConfigMap(viper2.AllSettings())
		if err != nil {
			return nil, errors.WithStackIf(err)
		}
	}
	return &Config{v: _viper_}, nil
}
// LoadAutoConfig creates a new empty Config, typically used when configuration
// is provided programmatically rather than from files.
func LoadAutoConfig() *Config {
	return NewConfig()
}

// NewFromBytes creates a Config from raw bytes with the specified format (json, yaml, toml, etc.)
func NewFromBytes(data []byte, format string) (*Config, error) {
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

	return &Config{v: v}, nil
}

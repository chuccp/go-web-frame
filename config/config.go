package config

import (
	"bytes"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"

	"emperror.dev/errors"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/go-viper/encoding/ini"
	"github.com/go-viper/mapstructure/v2"
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
	AllSettings() map[string]any
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
// Supports both camelCase and snake_case keys in config files.
func (c *Config) UnmarshalKey(key string, v any) error {
	return errors.WithStackIf(c.v.UnmarshalKey(key, v, decoderOpt))
}

// Unmarshal unmarshals the entire configuration into the target struct.
// Supports both camelCase and snake_case keys in config files.
func (c *Config) Unmarshal(v any) error {
	return errors.WithStackIf(c.v.Unmarshal(v, decoderOpt))
}

// decoderOpt is a viper decoder option that normalizes map keys to support
// both camelCase and snake_case config keys transparently.
var decoderOpt = viper.DecodeHook(normalizeKeysHook)

// normalizeKeysHook is a mapstructure decode hook that, for every map→struct
// decode, duplicates each map key into both camelCase and snake_case forms.
// This allows config files to use either naming convention without struct tags.
var normalizeKeysHook = mapstructure.DecodeHookFunc(
	func(from, to reflect.Type, data any) (any, error) {
		if from.Kind() == reflect.Map && to.Kind() == reflect.Struct {
			if m, ok := data.(map[string]any); ok {
				return addKeyVariants(m), nil
			}
		}
		return data, nil
	},
)

// addKeyVariants returns a new map where each key also has a counterpart in the
// other naming convention: snake_case ↔ camelCase. Nested maps are processed
// recursively. If both forms already exist, the original is kept.
func addKeyVariants(m map[string]any) map[string]any {
	out := make(map[string]any, len(m)*2)
	for k, v := range m {
		out[k] = v
		if sub, ok := v.(map[string]any); ok {
			v = addKeyVariants(sub)
			out[k] = v
		}
		alt := alternateKey(k)
		if alt != k {
			if _, exists := out[alt]; !exists {
				out[alt] = v
			}
		}
	}
	return out
}

// alternateKey converts a snake_case key to camelCase or vice versa.
func alternateKey(s string) string {
	if strings.Contains(s, "_") {
		return snakeToCamel(s)
	}
	return camelToSnake(s)
}

func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) == 1 {
		return s
	}
	var b strings.Builder
	for i, p := range parts {
		if p == "" {
			continue
		}
		if i == 0 {
			b.WriteString(strings.ToLower(p))
		} else {
			runes := []rune(strings.ToLower(p))
			if len(runes) > 0 {
				runes[0] = unicode.ToUpper(runes[0])
			}
			b.WriteString(string(runes))
		}
	}
	return b.String()
}

func camelToSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
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

// AllSettings returns all configuration settings as a map.
func (c *Config) AllSettings() map[string]any {
	return c.v.AllSettings()
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

// MergeConfig merges multiple IConfig instances into a single *Config.
// Later configs take precedence over earlier ones for duplicate keys.
func MergeConfig(configs ...IConfig) *Config {
	target := viper.New()
	for _, cfg := range configs {
		_ = target.MergeConfigMap(cfg.AllSettings())
	}
	return &Config{v: target}
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

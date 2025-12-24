package config

import (
	"path/filepath"

	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"github.com/go-viper/encoding/ini"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type IConfig interface {
	GetString(key string) string
	Put(key string, value any)
	HasKey(key string) bool
	GetStringOrDefault(key string, defaultValue string) string
	GetInt(key string) int
	GetIntOrDefault(key string, defaultValue int) int
	GetBoolOrDefault(key string, defaultValue bool) bool
	Unmarshal(key string, v any) error
	ReplaceKey(key string, newKey string)
}

type Config struct {
	v *viper.Viper
}

func (c *Config) GetString(key string) string {
	return c.v.GetString(key)
}
func (c *Config) Put(key string, value any) {
	c.v.Set(key, value)
}
func (c *Config) GetStringOrDefault(key string, defaultValue string) string {
	v := c.v.GetString(key)
	if util.IsBlank(v) {
		return defaultValue
	}
	return v
}
func (c *Config) HasKey(key string) bool {
	return c.v.IsSet(key)
}
func (c *Config) Unmarshal(key string, v any) error {
	return log.WrapError(c.v.UnmarshalKey(key, v))
}

func (c *Config) GetInt(key string) int {
	return c.v.GetInt(key)
}

func (c *Config) GetIntOrDefault(key string, defaultValue int) int {
	v := c.v.GetInt(key)
	if v == 0 {
		return defaultValue
	}
	return v
}
func (c *Config) GetBoolOrDefault(key string, defaultValue bool) bool {
	if util.IsBlank(key) {
		return defaultValue
	}
	return c.v.GetBool(key)
}
func (c *Config) ReplaceKey(key string, newKey string) {
	if c.v.IsSet(key) {
		c.v.Set(newKey, c.v.Get(key))
	}
}

type SingleFileConfig struct {
	*Config
	path string
}

func (c *SingleFileConfig) WriteConfig() error {
	return c.v.WriteConfig()
}
func LoadSingleFileConfig(path string) (*SingleFileConfig, error) {
	registry := viper.NewCodecRegistry()
	er := registry.RegisterCodec("ini", ini.Codec{})
	if er != nil {
		return nil, er
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return &SingleFileConfig{Config: &Config{v: _viper_}, path: absPath}, nil
}

func NewConfig() *Config {
	return &Config{v: viper.New()}
}
func LoadConfig(paths ...string) (*Config, error) {
	registry := viper.NewCodecRegistry()
	er := registry.RegisterCodec("ini", ini.Codec{})
	if er != nil {
		return nil, er
	}
	_viper_ := viper.New()
	for _, path := range paths {
		viper2 := viper.NewWithOptions(viper.WithCodecRegistry(registry))
		viper2.SetConfigFile(path)
		err := viper2.ReadInConfig()
		if err != nil {
			return nil, err
		}
		err = _viper_.MergeConfigMap(viper2.AllSettings())
		if err != nil {
			return nil, err
		}
	}
	return &Config{v: _viper_}, nil
}
func LoadAutoConfig() *Config {
	return NewConfig()
}

package config

import (
	"bytes"
	"log"
	"path/filepath"

	"github.com/chuccp/go-web-frame/util"
	"github.com/spf13/viper"
)

type Config struct {
	v *viper.Viper
}

func (c *Config) GetString(key string) string {
	return c.v.GetString(key)
}

func (c *Config) GetStringOrDefault(key string, defaultValue string) string {
	v := c.v.GetString(key)
	if util.IsBlank(v) {
		return defaultValue
	}
	return v
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
func LoadConfig(paths ...string) (*Config, error) {

	_viper_ := viper.New()
	_viper_.SetConfigType("yaml")
	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		log.Printf("加载配置文件 %v", absPath)
		data, err := util.ReadFileBytes(absPath)
		if err != nil {
			return nil, err
		}
		err = _viper_.MergeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
	}
	return &Config{v: _viper_}, nil
}

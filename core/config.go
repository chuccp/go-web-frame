package core

import (
	"bytes"

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
func LoadConfig(path string) (*Config, error) {

	yml, err := util.ReadFileBytes(path)
	if err != nil {
		return nil, err
	}
	_viper_ := viper.New()
	_viper_.SetConfigType("yaml")
	err = _viper_.ReadConfig(bytes.NewBuffer(yml))
	if err != nil {
		return nil, err
	}
	return &Config{v: _viper_}, nil
}

package wf

import config2 "github.com/chuccp/go-web-frame/config"

func LoadConfig(paths ...string) (*config2.Config, error) {
	return config2.LoadConfig(paths...)
}
func LoadAutoConfig() *config2.Config {
	return config2.LoadAutoConfig()
}

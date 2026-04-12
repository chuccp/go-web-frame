package wf

import (
	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
)

func LoadConfig(paths ...string) (*config2.Config, error) {
	return config2.LoadConfig(paths...)
}
func LoadAutoConfig() *config2.Config {
	return config2.LoadAutoConfig()
}

func defaultRestGroup(serverConfig *web.ServerConfig, handles *web.Handles) *core.RestGroup {
	return core.NewRestGroupBuilder().ServerConfig(serverConfig).Handles(handles).Build()
}

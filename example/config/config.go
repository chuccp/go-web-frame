package main

import (
	"context"
	"fmt"

	wf "github.com/chuccp/go-web-frame"
	"github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
	"go.uber.org/zap"
)

// AppConfig represents the application configuration structure
type AppConfig struct {
	Server ServerConfig `mapstructure:"server"`
	Web    WebConfig    `mapstructure:"web"`
	Log    LogConfig    `mapstructure:"log"`
	Redis  RedisConfig  `mapstructure:"redis"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type WebConfig struct {
	DB DatabaseConfig `mapstructure:"db"`
}

type DatabaseConfig struct {
	Type             string `mapstructure:"type"`
	Host             string `mapstructure:"host"`
	Port             int    `mapstructure:"port"`
	Username         string `mapstructure:"username"`
	Password         string `mapstructure:"password"`
	Database         string `mapstructure:"database"`
	MaxOpenConns     int    `mapstructure:"max_open_conns"`
	MaxIdleConns     int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime  int    `mapstructure:"conn_max_lifetime"`
}

type LogConfig struct {
	Level     string `mapstructure:"level"`
	Path      string `mapstructure:"path"`
	MaxSize   int    `mapstructure:"max_size"`
	MaxBackups int   `mapstructure:"max_backups"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// ConfigService demonstrates how to use configuration in a service
type ConfigService struct {
	core.IService
	cfg *AppConfig
}

func (s *ConfigService) Init(ctx *core.Context) error {
	// Get the configuration from context
	cfg := ctx.GetConfig()

	// Unmarshal entire config into struct
	s.cfg = &AppConfig{}
	if err := cfg.Unmarshal(s.cfg); err != nil {
		log.Error("Failed to unmarshal config", zap.Error(err))
		return err
	}

	// Log configuration values
	log.Info("Application configuration loaded",
		zap.Int("port", s.cfg.Server.Port),
		zap.String("mode", s.cfg.Server.Mode),
		zap.String("db_type", s.cfg.Web.DB.Type),
		zap.String("db_host", s.cfg.Web.DB.Host),
		zap.String("log_level", s.cfg.Log.Level),
	)

	return nil
}

func (s *ConfigService) GetServerPort() int {
	return s.cfg.Server.Port
}

func (s *ConfigService) GetDBConfig() *DatabaseConfig {
	return &s.cfg.Web.DB
}

// ConfigController handles HTTP endpoints that expose configuration
type ConfigController struct {
	core.IRest
	service *ConfigService
}

func (c *ConfigController) Init(ctx *core.Context) error {
	// Get the config service from context
	c.service = core.GetService[*ConfigService](ctx)

	// Register routes
	ctx.Get("/config", c.GetConfig)
	ctx.Get("/config/server", c.GetServerConfig)
	ctx.Get("/config/database", c.GetDatabaseConfig)
	ctx.Get("/config/redis", c.GetRedisConfig)

	return nil
}

func (c *ConfigController) GetConfig(req *web.Request) (any, error) {
	return map[string]any{
		"server": c.service.cfg.Server,
		"web":    c.service.cfg.Web,
		"log":    c.service.cfg.Log,
		"redis":  c.service.cfg.Redis,
	}, nil
}

func (c *ConfigController) GetServerConfig(req *web.Request) (any, error) {
	return c.service.cfg.Server, nil
}

func (c *ConfigController) GetDatabaseConfig(req *web.Request) (any, error) {
	return c.service.cfg.Web.DB, nil
}

func (c *ConfigController) GetRedisConfig(req *web.Request) (any, error) {
	return c.service.cfg.Redis, nil
}

// Example of loading config from a single file
func loadConfigFromFile(filePath string) (*config.SingleFileConfig, error) {
	cfg, err := config.LoadSingleFileConfig(filePath)
	if err != nil {
		return nil, err
	}

	// Example: Read specific values
	port := cfg.GetInt("server.port")
	mode := cfg.GetString("server.mode")
	dbHost := cfg.GetString("web.db.host")

	fmt.Printf("Loaded config from file:\n")
	fmt.Printf("  Port: %d\n", port)
	fmt.Printf("  Mode: %s\n", mode)
	fmt.Printf("  DB Host: %s\n", dbHost)

	return cfg, nil
}

// Example of loading config from multiple files
func loadConfigFromMultipleFiles(files ...string) (*config.Config, error) {
	cfg, err := config.LoadConfig(files...)
	if err != nil {
		return nil, err
	}

	// Example: Check if key exists and get default value
	if cfg.HasKey("server.ssl.enabled") {
		sslEnabled := cfg.GetBoolOrDefault("server.ssl.enabled", false)
		fmt.Printf("SSL Enabled: %v\n", sslEnabled)
	}

	return cfg, nil
}

// Example of creating config from bytes
func loadConfigFromBytes() (*config.Config, error) {
	yamlData := []byte(`
server:
  port: 9090
  mode: debug

log:
  level: debug
  path: /tmp/app.log
`)

	cfg, err := config.NewFromBytes(yamlData, "yaml")
	if err != nil {
		return nil, err
	}

	fmt.Printf("Loaded config from bytes:\n")
	fmt.Printf("  Port: %d\n", cfg.GetInt("server.port"))
	fmt.Printf("  Log Path: %s\n", cfg.GetString("log.path"))

	return cfg, nil
}

func main() {
	// Example 1: Load config from a single file
	// Uncomment to test:
	// cfg, err := loadConfigFromFile("./application.example.yml")
	// if err != nil {
	// 	log.Error("Failed to load config file", err)
	// 	return
	// }

	// Example 2: Load config from bytes
	// Uncomment to test:
	// _, err = loadConfigFromBytes()
	// if err != nil {
	// 	log.Error("Failed to load config from bytes", err)
	// }

	// Main application using auto config
	app := wf.NewWithAutoConfig()

	// Add config service
	app.AddService(&ConfigService{})

	// Add config controller
	app.AddRest(&ConfigController{})

	// Add a simple health check endpoint
	app.Get("/health", func(c *web.Request) (any, error) {
		return map[string]string{
			"status": "healthy",
			"config": "loaded",
		}, nil
	})

	// Add a custom config value (demonstrating Put method)
	app.AddService(&CustomConfigService{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := app.Run(ctx); err != nil {
		log.PrintPanic(err)
	}
}

// CustomConfigService demonstrates dynamic config manipulation
type CustomConfigService struct {
	core.IService
}

func (s *CustomConfigService) Init(ctx *core.Context) error {
	cfg := ctx.GetConfig()

	// Set custom configuration values
	cfg.Put("app.name", "Go Web Frame Example")
	cfg.Put("app.version", "1.0.0")
	cfg.Put("app.features.cache_enabled", true)

	// Read back the values
	appName := cfg.GetString("app.name")
	appVersion := cfg.GetString("app.version")
	cacheEnabled := cfg.GetBoolOrDefault("app.features.cache_enabled", false)

	log.Info("Custom configuration set",
		zap.String("app_name", appName),
		zap.String("app_version", appVersion),
		zap.Bool("cache_enabled", cacheEnabled),
	)

	return nil
}

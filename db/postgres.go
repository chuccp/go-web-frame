package db

import (
	"fmt"

	"emperror.dev/errors"
	log2 "github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/util"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresConfig holds PostgreSQL connection configuration
type PostgresConfig struct {
	Host     string
	Port     int
	Username string
	User     string
	Password string
	Database string
	Dbname   string
	SSLMode  string // disable, require, verify-ca, verify-full

	// Connection pool settings
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // in seconds
}

// GetMaxOpenConns implements ConnectionPoolConfig
func (c *PostgresConfig) GetMaxOpenConns() int {
	if c.MaxOpenConns == 0 {
		return 100
	}
	return c.MaxOpenConns
}

// GetMaxIdleConns implements ConnectionPoolConfig
func (c *PostgresConfig) GetMaxIdleConns() int {
	if c.MaxIdleConns == 0 {
		return 10
	}
	return c.MaxIdleConns
}

// GetConnMaxLifetime implements ConnectionPoolConfig
func (c *PostgresConfig) GetConnMaxLifetime() int {
	if c.ConnMaxLifetime == 0 {
		return 3600 // 1 hour
	}
	return c.ConnMaxLifetime
}

// Connection creates a new PostgreSQL database connection
func (c *PostgresConfig) Connection() (db *DB, err error) {
	if util.IsBlank(c.Username) {
		c.Username = c.User
	}
	if util.IsBlank(c.Database) {
		c.Database = c.Dbname
	}
	if c.Port == 0 {
		c.Port = 5432
	}
	if util.IsBlank(c.SSLMode) {
		c.SSLMode = "disable" // default to disable for local development
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode)

	log2.Debug("postgres", zap.String("dsn", fmt.Sprintf("host=%s port=%d user=%s password=****** dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Database, c.SSLMode)))

	db_, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, errors.WithStackIf(err)
	}
	if err := db_.Use(&ZeroTimePlugin{}); err != nil {
		return nil, errors.WithStackIf(err)
	}

	if err := ApplyConnectionPool(db_, c); err != nil {
		return nil, err
	}

	return &DB{db: db_}, nil
}

// ConnectionPostgres creates a new PostgreSQL connection with given parameters
func ConnectionPostgres(host string, port int, username, password, database string, sslMode string) (db *DB, err error) {
	var pgConfig = &PostgresConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		Database: database,
		SSLMode:  sslMode,
	}
	return pgConfig.Connection()
}

package config

import (
	"fmt"

	"github.com/shirr9/pr-reviewer-service/internal/infrastructure/logger"
)

// Config represents the application configuration.
type Config struct {
	Env        string `yaml:"env" env-default:"local"`
	PostgresDb `yaml:"postgres"`
}

// Validate checks the correctness of the configuration.
func (c *Config) Validate() error {
	if c.Env != logger.EnvDev && c.Env != logger.EnvLocal && c.Env != logger.EnvProd {
		return fmt.Errorf("invalid env: %s (must be one of: %s, %s, %s)",
			c.Env, logger.EnvDev, logger.EnvLocal, logger.EnvProd)
	}
	return nil
}

// PostgresDb contains PostgreSQL database connection parameters.
type PostgresDb struct {
	Username string `yaml:"user"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD" env-required:"true"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	DbName   string `yaml:"db_name"`
	SSlMode  string `yaml:"sslmode" env-default:"disable"`
}

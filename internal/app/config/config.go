package config

import "time"

// Config represents the application configuration.
type Config struct {
	Env        string     `yaml:"env" env-default:"local"`
	Server     Server     `yaml:"server"`
	PostgresDb PostgresDb `yaml:"postgres"`
}

// Server contains HTTP server configuration.
type Server struct {
	Port         int           `yaml:"port" env-default:"8080"`
	Env          string        `yaml:"env" env-default:"local"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env-default:"10s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env-default:"10s"`
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

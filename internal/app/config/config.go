package config

// Config represents the application configuration.
type Config struct {
	Env        string `yaml:"env" env-default:"local"`
	PostgresDb `yaml:"postgres"`
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

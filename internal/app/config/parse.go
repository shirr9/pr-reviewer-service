package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// MustLoad loads application configuration from the specified YAML file.
// It attempts to load environment variables from .env file (optional for Docker).
func MustLoad(path string) (*Config, error) {
	// Load .env if exists (optional for Docker)
	_ = godotenv.Load()

	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("config file not found at %s: %w", path, err)
		}
		return nil, err
	}

	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

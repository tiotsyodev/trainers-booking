package postgres

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host     string        `required:"true"`
	Port     int           `default:"5432"`
	User     string        `required:"true"`
	Password string        `required:"true"`
	Name     string        `required:"true"`
	Timeout  time.Duration `default:"30s"`
}

func loadConfig() (Config, error) {
	var cfg Config
	if err := envconfig.Process("DB", &cfg); err != nil {
		return Config{}, fmt.Errorf("envconfig process: %w", err)
	}

	return cfg, nil
}

func ConfigMustLoad() Config {
	config, err := loadConfig()
	if err != nil {
		panic(err)
	}

	return config
}

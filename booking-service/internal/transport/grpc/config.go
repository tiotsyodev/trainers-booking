package grpc

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port string `envconfig:"GRPC_PORT" ,default:"50501"`
}

func loadConfig() (Config, error) {

	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return cfg, fmt.Errorf("envconfig process: %w", err)
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

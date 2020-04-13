package config

import (
	"log"

	"github.com/caarlos0/env"
)

// Config use for configuration app params
type Config struct {
	Listen   string `env:"LISTEN" envDefault:"localhost:8080"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"debug"`
}

// Load - set config from env vareiables
func Load() Config {
	cfg := Config{}
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	return cfg
}

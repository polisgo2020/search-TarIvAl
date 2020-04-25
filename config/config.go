package config

import (
	"log"

	"github.com/caarlos0/env"
)

// Config use for configuration app params
type Config struct {
	Listen   string `env:"LISTEN" envDefault:"localhost:8080"`
	PgSQL    string `env:"PGSQL" envDefault:"postgres://postgres:111111@localhost:5432/index?sslmode=disable"`
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

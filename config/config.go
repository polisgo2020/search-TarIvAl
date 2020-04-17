package config

import (
	"log"

	"github.com/caarlos0/env"
)

// Config use for configuration app params
type Config struct {
	Listen   string `env:"LISTEN" envDefault:"localhost:8080"`
	PgSQL    string `env:"PGSQL" envDefault:"host=localhost port=5432 user=postgres password=111111 dbname=index sslmode=disable"`
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

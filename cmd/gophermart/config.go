package main

import (
	"flag"

	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog/log"
)

type Config struct {
	Address     string `env:"RUN_ADDRESS"`
	DatabaseURI string `env:"DATABASE_URI"`
	Accrual     string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	AuthKey     string `env:"Auth Key"`
}

func ReadConfig() Config {
	var config Config

	flag.StringVar(&config.Address, "a", "127.0.0.1:8000", "Server Address")
	flag.StringVar(&config.DatabaseURI, "d", "", "Database URI")
	flag.StringVar(&config.Accrual, "r", "", "Accrual address")
	flag.StringVar(&config.AuthKey, "k", "", "Auth key")
	flag.Parse()

	if err := env.Parse(&config); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config")
	}

	return config
}

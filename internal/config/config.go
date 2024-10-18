package config

import "os"

type Config struct {
	Address string
}

func Load() (*Config, error) {
	address := os.Getenv("SERVER_ADDRESS")
	if address == "" {
		address = ":8080"
	}

	return &Config{
		Address: address,
	}, nil
}

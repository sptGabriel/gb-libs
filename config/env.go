package config

import (
	"fmt"

	"github.com/joho/godotenv"
)

func LoadEnv(file string) error {
	if file == "" {
		file = ".env"
	}

	if err := godotenv.Load(file); err != nil {
		return fmt.Errorf("on load envfile: %v", err)
	}

	return nil
}

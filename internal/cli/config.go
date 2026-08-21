package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func loadDotEnv(path string) error {
	err := godotenv.Load(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return err
}

func requiredEnv(name, hint string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("%s is required (%s)", name, hint)
	}

	return value, nil
}

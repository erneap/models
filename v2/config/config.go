package config

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func Config(key string) string {
	answer := strings.TrimSpace(os.Getenv((key)))
	if answer == "" {
		exists, err := FileExists(".env")
		if err != nil {
			log.Println(err)
			return ""
		}
		if !exists {
			return ""
		}
		err = godotenv.Load(".env")
		if err != nil {
			log.Print("Error loading .env file")
			log.Println(err.Error())
		}
	}

	return strings.TrimSpace(os.Getenv(key))
}

func FileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

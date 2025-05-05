package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func Config(key string) string {
	answer := strings.TrimSpace(os.Getenv((key)))
	if answer == "" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Print("Error loading .env file")
			log.Println(err.Error())
		}
	}

	return strings.TrimSpace(os.Getenv(key))
}

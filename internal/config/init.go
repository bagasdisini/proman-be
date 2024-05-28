package config

import (
	"github.com/joho/godotenv"
	"proman-backend/pkg/log"
)

func InitConfig(filepath string) {
	err := godotenv.Load(filepath)
	if err != nil {
		log.Info(".env file not found, using environment variables instead.")
	}

	initApp()
	initAws()
	initMongo()
}

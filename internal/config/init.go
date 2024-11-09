package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"path"
)

func init() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	filepath := path.Join(dir, ".env")
	err = godotenv.Load(filepath)
	if err != nil {
		fmt.Println(".env file not found, using environment variables instead.")
	}

	initApp()
	initAws()
	initMongo()
	initGitlab()
}

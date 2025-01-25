package config

import "os"

var Mongo struct {
	Url  string `mapstructure:"MONGODB_URI"`
	Name string `mapstructure:"MONGODB_NAME"`
}

func initMongo() {
	Mongo.Url = os.Getenv("MONGODB_URI")
	Mongo.Name = os.Getenv("MONGODB_NAME")

	if Mongo.Url == "" {
		panic("MONGODB_URI is not set")
	}
	if Mongo.Name == "" {
		panic("MONGODB_NAME is not set")
	}
}

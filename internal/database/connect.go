package database

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"proman-backend/config"
)

var db *mongo.Database

func ConnectMongo() *mongo.Database {
	if db == nil {
		client, err := mongo.Connect(options.Client().ApplyURI(config.Mongo.Url))
		if err != nil {
			panic(err)
		}

		if err := client.Ping(context.Background(), nil); err != nil {
			panic("Failed to connect to MongoDB: " + err.Error())
		}

		db = client.Database(config.Mongo.Name)
	}
	return db
}

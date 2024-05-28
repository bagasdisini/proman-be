package database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"proman-backend/internal/config"
)

var db *mongo.Database

func ConnectMongo() *mongo.Database {
	if db == nil {
		client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(config.Mongo.Url))
		if err != nil {
			panic(err)
		}

		db = client.Database(config.Mongo.Name)
	}
	return db
}

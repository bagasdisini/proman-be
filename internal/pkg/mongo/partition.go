package mongo

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"sync"
)

type PartitionRepository struct {
	prefixName  string
	db          *mongo.Database
	collections sync.Map
}

func NewPartitionRepository(db *mongo.Database, prefixName string) *PartitionRepository {
	return &PartitionRepository{
		prefixName:  prefixName,
		db:          db,
		collections: sync.Map{},
	}
}

func (r *PartitionRepository) Coll(partitionId bson.ObjectID) *mongo.Collection {
	hex := partitionId.Hex()
	if coll, ok := r.collections.Load(hex); ok {
		return coll.(*mongo.Collection)
	}
	coll := r.db.Collection(r.prefixName + "_" + hex)
	r.collections.Store(hex, coll)
	return coll
}

package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PaginationResult struct {
	Total  int64       `json:"total"`
	Offset int         `json:"offset"`
	Limit  int         `json:"limit"`
	List   interface{} `json:"list"`
}

func BuildPaginateOrderOptionByField(sortParam bson.D, offset, limit int) (*options.FindOptions, error) {
	if offset < 0 {
		offset = 0
	}

	if limit <= 0 {
		limit = 10
	}

	findOptions := options.Find()
	findOptions.SetSort(sortParam)
	findOptions.SetSkip(int64(offset))
	findOptions.SetLimit(int64(limit))
	return findOptions, nil
}

package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math"
)

type PaginationResult struct {
	Result interface{} `json:"result"`
	Total  int         `json:"total"`
	Pages  int         `json:"pages"`
	Page   int         `json:"page"`
	Limit  int         `json:"limit"`
}

func BuildPaginateOrderOptionByField(sortParam bson.D, page, limit int) (*options.FindOptions, error) {
	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 10
	}

	skipped := (page - 1) * limit
	findOptions := options.Find()
	findOptions.SetSort(sortParam)
	findOptions.SetSkip(int64(skipped))
	findOptions.SetLimit(int64(limit))

	return findOptions, nil
}

func CalculateTotalPages(totalData, limit, page int) (int, bool) {
	totalPages := int(math.Ceil(float64(totalData) / float64(limit)))
	pageOutOfRange := page > totalPages && totalPages != 0
	return totalPages, pageOutOfRange
}

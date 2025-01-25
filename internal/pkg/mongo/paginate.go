package mongo

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"math"
)

type PaginationResult struct {
	Result interface{} `json:"result"`
	Total  int64       `json:"total"`
	Page   int64       `json:"page"`
	Pages  int64       `json:"pages"`
	Limit  int64       `json:"limit"`
}

func BuildPaginateOrderOptionByField(sortParam bson.D, page, limit int64) (*options.FindOptionsBuilder, error) {
	if page <= 0 {
		page = 1
	}

	if limit <= 0 {
		limit = 10
	}

	skipped := (page - 1) * limit
	findOptions := options.Find()
	findOptions.SetSort(sortParam)
	findOptions.SetSkip(skipped)
	findOptions.SetLimit(limit)

	return findOptions, nil
}

func calculateTotalPages(totalData, limit, page int64) (int64, bool) {
	return int64(math.Ceil(float64(totalData) / float64(limit))),
		page > int64(math.Ceil(float64(totalData)/float64(limit))) && int(math.Ceil(float64(totalData)/float64(limit))) != 0
}

func MakeResult(data interface{}, totalData int64, page, limit int64) *PaginationResult {
	totalPages, pageOutOfRange := calculateTotalPages(totalData, limit, page)
	if pageOutOfRange {
		return &PaginationResult{
			Result: nil,
			Total:  totalData,
			Pages:  totalPages,
			Page:   page,
			Limit:  limit,
		}
	}

	return &PaginationResult{
		Result: data,
		Total:  totalData,
		Pages:  totalPages,
		Page:   page,
		Limit:  limit,
	}
}

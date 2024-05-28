package util

import "go.mongodb.org/mongo-driver/bson"

func AddFilter(filter, newFilter bson.M) bson.M {
	if _, ok := filter["$and"]; ok {
		filter["$and"] = append(filter["$and"].([]bson.M), newFilter)
	} else {
		filter["$and"] = []bson.M{newFilter}
	}
	return filter
}

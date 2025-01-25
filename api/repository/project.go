package repository

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"proman-backend/config"
	"proman-backend/internal/pkg/const"
	"proman-backend/internal/pkg/util"
	"strings"
	"time"
)

type Project struct {
	ID          bson.ObjectID   `json:"_id" bson:"_id"`
	Name        string          `json:"name" bson:"name"`
	Description string          `json:"description" bson:"description"`
	Type        string          `json:"type" bson:"type"`
	StartDate   time.Time       `json:"start_date" bson:"start_date"`
	EndDate     time.Time       `json:"end_date" bson:"end_date"`
	Contributor []bson.ObjectID `json:"contributor" bson:"contributor"`
	Attachments []string        `json:"attachments" bson:"attachments"`
	Status      string          `json:"status" bson:"status"` // active, completed, pending, cancelled
	Logo        string          `json:"logo" bson:"logo"`
	CreatedAt   time.Time       `json:"created_at" bson:"created_at"`
	IsDeleted   bool            `json:"-" bson:"is_deleted"`
	TaskCount   CountTaskDetail `json:"task_count" bson:"task_count"`
}

func (u *Project) MarshalJSON() ([]byte, error) {
	type Alias Project
	var url string
	if u.Logo != "" {
		url = "https://" + config.S3.Bucket
		if !strings.Contains(config.S3.EndPoint, "https://") {
			url = url + "." + config.S3.EndPoint + "/" + u.Logo
		} else {
			url = url + "." + config.S3.EndPoint[8:] + "/" + u.Logo
		}
	}
	var attachments []string
	if u.Attachments != nil {
		for _, attachment := range u.Attachments {
			urls := "https://" + config.S3.Bucket
			if !strings.Contains(config.S3.EndPoint, "https://") {
				urls = urls + "." + config.S3.EndPoint + "/" + attachment
			} else {
				urls = urls + "." + config.S3.EndPoint[8:] + "/" + attachment
			}
			attachments = append(attachments, urls)
		}
	}
	return json.Marshal(&struct {
		*Alias
		Attachments []string `json:"attachments" bson:"attachments"`
		Logo        string   `json:"logo" bson:"logo"`
	}{
		Alias:       (*Alias)(u),
		Attachments: attachments,
		Logo:        url,
	})
}

type CountProjectDetail struct {
	Total     int `json:"total"`
	Active    int `json:"active"`
	Completed int `json:"completed"`
	Pending   int `json:"pending"`
	Cancelled int `json:"cancelled"`
}

type CountTypeProject struct {
	Type  string `json:"type"`
	Total int    `json:"total"`
}

type ProjectCollRepository struct {
	coll *mongo.Collection
}

func NewProjectRepository(db *mongo.Database) *ProjectCollRepository {
	return &ProjectCollRepository{
		coll: db.Collection("projects"),
	}
}

func (r *ProjectCollRepository) FindAll(cq *util.CommonQuery) (*[]Project, error) {
	var projects []Project

	matchStage := bson.M{"is_deleted": bson.M{"$ne": true}}

	if len(cq.Q) > 0 {
		matchStage["$or"] = []bson.M{
			{"name": bson.M{"$regex": bson.Regex{Pattern: cq.Q, Options: "i"}}},
			{"description": bson.M{"$regex": bson.Regex{Pattern: cq.Q, Options: "i"}}},
		}
	}

	if len(cq.Status) > 0 && _const.IsValidProjectStatus(cq.Status) {
		matchStage["status"] = cq.Status
	}

	if cq.UserId != bson.NilObjectID {
		matchStage["contributor"] = cq.UserId
	}

	if existingOr, ok := matchStage["$or"]; ok {
		matchStage["$or"] = append(existingOr.([]bson.M), bson.M{
			"$or": []bson.M{
				{
					"start_date": bson.M{"$lt": cq.End},
					"end_date":   bson.M{"$gte": cq.Start},
				},
				{
					"start_date": bson.M{"$gte": cq.Start, "$lt": cq.End},
				},
			},
		})
	} else {
		matchStage["$or"] = []bson.M{
			{
				"start_date": bson.M{"$lt": cq.End},
				"end_date":   bson.M{"$gte": cq.Start},
			},
			{
				"start_date": bson.M{"$gte": cq.Start, "$lt": cq.End},
			},
		}
	}

	skip := (cq.Page - 1) * cq.Limit
	limit := cq.Limit

	pipeline := []bson.M{
		{
			"$match": matchStage,
		},
		{
			"$lookup": bson.M{
				"from":         "tasks",
				"localField":   "_id",
				"foreignField": "project_id",
				"as":           "tasks",
			},
		},
		{
			"$addFields": bson.M{
				"task_count": bson.M{
					"active": bson.M{
						"$size": bson.M{
							"$filter": bson.M{
								"input": "$tasks",
								"as":    "task",
								"cond":  bson.M{"$eq": []interface{}{"$$task.status", _const.TaskActive}},
							},
						},
					},
					"testing": bson.M{
						"$size": bson.M{
							"$filter": bson.M{
								"input": "$tasks",
								"as":    "task",
								"cond":  bson.M{"$eq": []interface{}{"$$task.status", _const.TaskTesting}},
							},
						},
					},
					"completed": bson.M{
						"$size": bson.M{
							"$filter": bson.M{
								"input": "$tasks",
								"as":    "task",
								"cond":  bson.M{"$eq": []interface{}{"$$task.status", _const.TaskCompleted}},
							},
						},
					},
					"cancelled": bson.M{
						"$size": bson.M{
							"$filter": bson.M{
								"input": "$tasks",
								"as":    "task",
								"cond":  bson.M{"$eq": []interface{}{"$$task.status", _const.TaskCancelled}},
							},
						},
					},
					"total": bson.M{"$size": "$tasks"},
				},
			},
		},
		{
			"$project": bson.M{
				"tasks": 0,
			},
		},
		{
			"$sort": bson.D{{"_id", cq.Sort}},
		},
		{
			"$skip": skip,
		},
		{
			"$limit": limit,
		},
	}

	cursor, err := r.coll.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(context.TODO(), &projects); err != nil {
		return nil, err
	}
	return &projects, nil
}

func (r *ProjectCollRepository) FindOneByID(_id bson.ObjectID) (*Project, error) {
	var project Project

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"_id":        _id,
				"is_deleted": bson.M{"$ne": true},
			},
		},
		{
			"$lookup": bson.M{
				"from":         "tasks",
				"localField":   "_id",
				"foreignField": "project_id",
				"as":           "tasks",
			},
		},
		{
			"$addFields": bson.M{
				"task_count": bson.M{
					"active": bson.M{
						"$size": bson.M{
							"$filter": bson.M{
								"input": "$tasks",
								"as":    "task",
								"cond":  bson.M{"$eq": []interface{}{"$$task.status", _const.TaskActive}},
							},
						},
					},
					"testing": bson.M{
						"$size": bson.M{
							"$filter": bson.M{
								"input": "$tasks",
								"as":    "task",
								"cond":  bson.M{"$eq": []interface{}{"$$task.status", _const.TaskTesting}},
							},
						},
					},
					"completed": bson.M{
						"$size": bson.M{
							"$filter": bson.M{
								"input": "$tasks",
								"as":    "task",
								"cond":  bson.M{"$eq": []interface{}{"$$task.status", _const.TaskCompleted}},
							},
						},
					},
					"cancelled": bson.M{
						"$size": bson.M{
							"$filter": bson.M{
								"input": "$tasks",
								"as":    "task",
								"cond":  bson.M{"$eq": []interface{}{"$$task.status", _const.TaskCancelled}},
							},
						},
					},
					"total": bson.M{"$size": "$tasks"},
				},
			},
		},
		{
			"$project": bson.M{
				"tasks": 0,
			},
		},
	}

	cursor, err := r.coll.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	if cursor.Next(context.TODO()) {
		if err := cursor.Decode(&project); err != nil {
			return nil, err
		}
	} else {
		return nil, mongo.ErrNoDocuments
	}
	return &project, nil
}

func (r *ProjectCollRepository) CountProject(cq *util.CommonQuery) (*CountProjectDetail, error) {
	var count []CountProjectDetail

	matchStage := bson.D{{"is_deleted", bson.M{"$ne": true}}}

	if len(cq.Q) > 0 {
		matchStage = append(matchStage, bson.E{Key: "$or", Value: bson.A{
			bson.D{{"name", bson.Regex{Pattern: cq.Q, Options: "i"}}},
			bson.D{{"description", bson.Regex{Pattern: cq.Q, Options: "i"}}},
		}})
	}

	if len(cq.Status) > 0 && _const.IsValidProjectStatus(cq.Status) {
		matchStage = append(matchStage, bson.E{Key: "status", Value: cq.Status})
	}

	if cq.UserId != bson.NilObjectID {
		matchStage = append(matchStage, bson.E{Key: "contributor", Value: cq.UserId})
	}

	matchStage = append(matchStage, bson.E{
		Key: "$or",
		Value: bson.A{
			bson.D{
				{"start_date", bson.M{"$lt": cq.End}},
				{"end_date", bson.M{"$gte": cq.Start}},
			},
			bson.D{
				{"start_date", bson.M{"$gte": cq.Start}},
				{"start_date", bson.M{"$lt": cq.End}},
			},
		},
	})

	match := bson.D{
		{"$match", matchStage},
	}

	group := bson.D{
		{"$group", bson.D{
			{"_id", nil},
			{"total", bson.D{{"$sum", 1}}},
			{_const.ProjectActive, bson.D{{"$sum", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$status", _const.ProjectActive}}}, 1, 0}}}}}},
			{_const.ProjectCompleted, bson.D{{"$sum", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$status", _const.ProjectCompleted}}}, 1, 0}}}}}},
			{_const.ProjectPending, bson.D{{"$sum", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$status", _const.ProjectPending}}}, 1, 0}}}}}},
			{_const.ProjectCancelled, bson.D{{"$sum", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$status", _const.ProjectCancelled}}}, 1, 0}}}}}},
		}},
	}

	cursor, err := r.coll.Aggregate(context.TODO(), mongo.Pipeline{match, group})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.TODO(), &count); err != nil {
		return nil, err
	}

	if len(count) > 0 {
		return &count[0], nil
	} else {
		return &CountProjectDetail{}, nil
	}
}

func (r *ProjectCollRepository) CountProjectTypes(cq *util.CommonQuery) (*[]CountTypeProject, error) {
	var count []CountTypeProject

	matchStage := bson.D{
		{"is_deleted", bson.M{"$ne": true}},
		{"status", bson.M{"$ne": _const.ProjectCancelled}},
	}

	if len(cq.Q) > 0 {
		matchStage = append(matchStage, bson.E{Key: "$or", Value: bson.A{
			bson.D{{"name", bson.Regex{Pattern: cq.Q, Options: "i"}}},
			bson.D{{"description", bson.Regex{Pattern: cq.Q, Options: "i"}}},
		}})
	}

	if len(cq.Status) > 0 && _const.IsValidProjectStatus(cq.Status) {
		matchStage = append(matchStage, bson.E{Key: "status", Value: cq.Status})
	}

	if cq.UserId != bson.NilObjectID {
		matchStage = append(matchStage, bson.E{Key: "contributor", Value: cq.UserId})
	}

	matchStage = append(matchStage, bson.E{
		Key: "$or",
		Value: bson.A{
			bson.D{
				{"start_date", bson.M{"$lt": cq.End}},
				{"end_date", bson.M{"$gte": cq.Start}},
			},
			bson.D{
				{"start_date", bson.M{"$gte": cq.Start}},
				{"start_date", bson.M{"$lt": cq.End}},
			},
		},
	})

	match := bson.D{
		{"$match", matchStage},
	}

	group := bson.D{
		{"$group", bson.D{
			{"_id", "$type"},
			{"type", bson.D{{"$first", "$type"}}},
			{"total", bson.D{{"$sum", 1}}},
		}},
	}

	cursor, err := r.coll.Aggregate(context.Background(), mongo.Pipeline{match, group})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.Background(), &count); err != nil {
		return nil, err
	}
	return &count, nil
}

func (r *ProjectCollRepository) InsertOne(projectData *Project) (*Project, error) {
	var data *Project
	dataInsert, err := r.coll.InsertOne(context.TODO(), projectData)
	if err != nil {
		return nil, err
	}

	insertedID := dataInsert.InsertedID.(bson.ObjectID)
	err = r.coll.FindOne(context.TODO(), bson.M{"_id": insertedID}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *ProjectCollRepository) DeleteOneByID(_id bson.ObjectID) error {
	_, err := r.coll.UpdateOne(context.TODO(), bson.M{"_id": _id}, bson.M{"$set": bson.M{"is_deleted": true}})
	if err != nil {
		return err
	}
	return nil
}

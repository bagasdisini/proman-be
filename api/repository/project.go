package repository

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"proman-backend/internal/config"
	"strings"
	"time"
)

type Project struct {
	ID          primitive.ObjectID   `json:"_id" bson:"_id"`
	Name        string               `json:"name" bson:"name"`
	Description string               `json:"description" bson:"description"`
	Type        string               `json:"type" bson:"type"`
	StartDate   time.Time            `json:"start_date" bson:"start_date"`
	EndDate     time.Time            `json:"end_date" bson:"end_date"`
	Contributor []primitive.ObjectID `json:"contributor" bson:"contributor"`
	Attachments []string             `json:"attachments" bson:"attachments"`
	Status      string               `json:"status" bson:"status"` // active, completed, pending
	Logo        string               `json:"logo" bson:"logo"`
	CreatedAt   time.Time            `json:"created_at" bson:"created_at"`
	IsDeleted   bool                 `json:"-" bson:"is_deleted"`
}

func (u *Project) MarshalJSON() ([]byte, error) {
	type Alias Project
	var url string
	if u.Logo != "" {
		url = "https://" + config.S3.Bucket
		if !strings.Contains(config.S3.EndPoint, "https://") {
			url = url + "." + config.S3.EndPoint + "/" + config.AWS.ProjectLogoDir + "/" + u.Logo
		} else {
			url = url + "." + config.S3.EndPoint[8:] + "/" + config.AWS.ProjectLogoDir + "/" + u.Logo
		}
	}
	return json.Marshal(&struct {
		*Alias
		Logo string `json:"logo" bson:"logo"`
	}{
		Alias: (*Alias)(u),
		Logo:  url,
	})
}

type CountProjectDetail struct {
	Total     int `json:"total"`
	Active    int `json:"active"`
	Completed int `json:"completed"`
	Pending   int `json:"pending"`
}

type CountProject struct {
	Current   CountProjectDetail `json:"current"`
	LastMonth CountProjectDetail `json:"last_month"`
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

func (r *ProjectCollRepository) FindOneByID(_id primitive.ObjectID) (*Project, error) {
	var user *Project
	filter := bson.M{
		"_id":        _id,
		"is_deleted": bson.M{"$ne": true},
	}

	err := r.coll.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *ProjectCollRepository) FindAllByContributorID(_id primitive.ObjectID) (*[]Project, error) {
	var projects []Project
	filter := bson.M{
		"contributor": _id,
		"is_deleted":  bson.M{"$ne": true},
	}

	cursor, err := r.coll.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(context.Background(), &projects); err != nil {
		return nil, err
	}
	return &projects, nil
}

func (r *ProjectCollRepository) CountProject(start, end time.Time) (*CountProjectDetail, error) {
	var count CountProjectDetail
	match := bson.D{
		{"$match", bson.D{
			{"is_deleted", bson.M{"$ne": true}},
			{"start_date", bson.M{"$gt": start}},
			{"end_date", bson.M{"$lt": end}},
		}},
	}
	group := bson.D{
		{"$group", bson.D{
			{"_id", nil},
			{"total", bson.D{{"$sum", 1}}},
			{"active", bson.D{{"$sum", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$status", "active"}}}, 1, 0}}}}}},
			{"completed", bson.D{{"$sum", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$status", "completed"}}}, 1, 0}}}}}},
			{"pending", bson.D{{"$sum", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$status", "pending"}}}, 1, 0}}}}}},
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

func (r *ProjectCollRepository) CountTypeProject() (*[]CountTypeProject, error) {
	var count []CountTypeProject
	match := bson.D{
		{"$match", bson.D{{"is_deleted", bson.M{"$ne": true}}}},
	}
	group := bson.D{
		{"$group", bson.D{
			{"_id", "$type"},
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

func (r *ProjectCollRepository) CountProjectsForUser(userID primitive.ObjectID) (int, error) {
	pipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"is_deleted", false},
			{"contributor", userID},
		}}},
		{{"$count", "total"}},
	}

	cursor, err := r.coll.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return 0, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			return
		}
	}(cursor, context.TODO())

	var result struct {
		Total int `bson:"total"`
	}
	if cursor.Next(context.TODO()) {
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
		return result.Total, nil
	}
	return 0, nil
}

func (r *ProjectCollRepository) InsertOne(projectData *Project) (*Project, error) {
	var data *Project
	dataInsert, err := r.coll.InsertOne(context.TODO(), projectData)
	if err != nil {
		return nil, err
	}

	insertedID := dataInsert.InsertedID.(primitive.ObjectID)
	err = r.coll.FindOne(context.TODO(), bson.M{"_id": insertedID}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

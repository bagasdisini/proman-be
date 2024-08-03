package repository

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"proman-backend/internal/config"
	_const "proman-backend/pkg/const"
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
	Status      string               `json:"status" bson:"status"` // active, completed, pending, cancelled
	Logo        string               `json:"logo" bson:"logo"`
	CreatedAt   time.Time            `json:"created_at" bson:"created_at"`
	IsDeleted   bool                 `json:"-" bson:"is_deleted"`
	TaskCount   CountTaskDetail      `json:"task_count" bson:"task_count"`
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

type CountProject struct {
	Current  CountProjectDetail `json:"current"`
	LastYear CountProjectDetail `json:"last_year"`
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

func (r *ProjectCollRepository) FindAll(taskRepo *TaskCollRepository) (*[]Project, error) {
	var projects []Project
	filter := bson.M{"is_deleted": bson.M{"$ne": true}}

	cursor, err := r.coll.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(context.Background(), &projects); err != nil {
		return nil, err
	}

	for i, project := range projects {
		count, err := taskRepo.FindAllByProjectID(project.ID)
		if err != nil {
			return nil, err
		}

		for _, c := range count {
			if c.Status == _const.TaskActive {
				projects[i].TaskCount.Active++
			} else if c.Status == _const.TaskTesting {
				projects[i].TaskCount.Testing++
			} else if c.Status == _const.TaskCompleted {
				projects[i].TaskCount.Completed++
			} else if c.Status == _const.TaskCancelled {
				projects[i].TaskCount.Cancelled++
			}
			projects[i].TaskCount.Total++
		}
	}
	return &projects, nil
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

func (r *ProjectCollRepository) CountProject(start, end time.Time) (*[]CountProjectDetail, error) {
	var count []CountProjectDetail
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
			{_const.ProjectActive, bson.D{{"$sum", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$status", _const.ProjectActive}}}, 1, 0}}}}}},
			{_const.ProjectCompleted, bson.D{{"$sum", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$status", _const.ProjectCompleted}}}, 1, 0}}}}}},
			{_const.ProjectPending, bson.D{{"$sum", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$status", _const.ProjectPending}}}, 1, 0}}}}}},
			{_const.ProjectCancelled, bson.D{{"$sum", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$status", _const.ProjectCancelled}}}, 1, 0}}}}}},
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
		{"$match", bson.D{
			{"is_deleted", bson.M{"$ne": true}},
			{"status", _const.ProjectActive},
		}},
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

func (r *ProjectCollRepository) DeleteOneByID(_id primitive.ObjectID) error {
	_, err := r.coll.UpdateOne(context.TODO(), bson.M{"_id": _id}, bson.M{"$set": bson.M{"is_deleted": true}})
	if err != nil {
		return err
	}
	return nil
}

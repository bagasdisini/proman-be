package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type Task struct {
	ID          primitive.ObjectID   `json:"_id" bson:"_id"`
	Name        string               `json:"name" bson:"name"`
	Description string               `json:"description" bson:"description"`
	StartDate   time.Time            `json:"start_date" bson:"start_date"`
	EndDate     time.Time            `json:"end_date" bson:"end_date"`
	Contributor []primitive.ObjectID `json:"contributor" bson:"contributor"`
	Status      string               `json:"status" bson:"status"` // active, testing, completed, cancelled
	ProjectID   primitive.ObjectID   `json:"project_id" bson:"project_id"`
	CreatedAt   time.Time            `json:"created_at" bson:"created_at"`
	IsDeleted   bool                 `json:"-" bson:"is_deleted"`
}

type TaskGroup struct {
	Active    []Task `json:"active"`
	Testing   []Task `json:"testing"`
	Completed []Task `json:"completed"`
	Cancelled []Task `json:"cancelled"`
}

type CountTaskDetail struct {
	Total     int `json:"total"`
	Active    int `json:"active"`
	Testing   int `json:"testing"`
	Completed int `json:"completed"`
	Cancelled int `json:"cancelled"`
}

type CountTask struct {
	Current   CountTaskDetail `json:"current"`
	LastMonth CountTaskDetail `json:"last_month"`
}

type CountUserActive struct {
	Total     int `json:"total"`
	Active    int `json:"active"`
	NotActive int `json:"not_active"`
}

type CountUser struct {
	Current   CountUserActive `json:"current"`
	LastMonth CountUserActive `json:"last_month"`
}

type TaskCollRepository struct {
	coll *mongo.Collection
}

func NewTaskRepository(db *mongo.Database) *TaskCollRepository {
	return &TaskCollRepository{
		coll: db.Collection("tasks"),
	}
}

func (r *TaskCollRepository) FindOneByID(_id primitive.ObjectID) (*Task, error) {
	var user *Task
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

func (r *TaskCollRepository) FindAllByProjectID(projectID primitive.ObjectID) ([]Task, error) {
	var tasks []Task
	filter := bson.M{
		"project_id": projectID,
		"is_deleted": bson.M{"$ne": true},
	}

	cursor, err := r.coll.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			return
		}
	}(cursor, context.TODO())

	for cursor.Next(context.TODO()) {
		var task Task
		if err := cursor.Decode(&task); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *TaskCollRepository) FindAllByUserID(userID primitive.ObjectID) ([]Task, error) {
	var tasks []Task
	filter := bson.M{
		"contributor": bson.M{"$in": []primitive.ObjectID{userID}},
		"is_deleted":  bson.M{"$ne": true},
	}

	cursor, err := r.coll.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			return
		}
	}(cursor, context.TODO())

	for cursor.Next(context.TODO()) {
		var task Task
		if err := cursor.Decode(&task); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *TaskCollRepository) FindAllByStatus(status string) ([]Task, error) {
	var tasks []Task
	filter := bson.M{
		"status":     status,
		"is_deleted": bson.M{"$ne": true},
	}

	cursor, err := r.coll.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			return
		}
	}(cursor, context.TODO())

	for cursor.Next(context.TODO()) {
		var task Task
		if err := cursor.Decode(&task); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *TaskCollRepository) FindAllGroupByStatus() (*TaskGroup, error) {
	var taskGroup TaskGroup
	match := bson.D{
		{"$match", bson.D{{"is_deleted", bson.M{"$ne": true}}}},
	}
	group := bson.D{
		{"$group", bson.D{
			{"_id", "$status"},
			{"tasks", bson.D{{"$push", "$$ROOT"}}},
		}},
	}

	cursor, err := r.coll.Aggregate(context.Background(), mongo.Pipeline{match, group})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.Background(), &taskGroup); err != nil {
		return nil, err
	}
	return &taskGroup, nil
}

func (r *TaskCollRepository) CountTask() (*CountTask, error) {
	var count CountTask
	match := bson.D{
		{"$match", bson.D{{"is_deleted", bson.M{"$ne": true}}}},
	}
	group := bson.D{
		{"$group", bson.D{
			{"_id", nil},
			{"total", bson.D{{"$sum", 1}}},
			{"active", bson.D{{"$sum", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$status", "active"}}}, 1, 0}}}}}},
			{"testing", bson.D{{"$sum", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$status", "testing"}}}, 1, 0}}}}}},
			{"completed", bson.D{{"$sum", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$status", "completed"}}}, 1, 0}}}}}},
			{"cancelled", bson.D{{"$sum", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$status", "cancelled"}}}, 1, 0}}}}}},
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

func (r *TaskCollRepository) CountActiveUsers(userRepo *UserCollRepository) (*CountUser, error) {
	var countUser CountUser

	taskPipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"is_deleted", false},
			{"status", bson.D{{"$in", bson.A{"active", "testing"}}}},
		}}},
		{{"$unwind", "$contributor"}},
		{{"$group", bson.D{
			{"_id", "$contributor"},
		}}},
	}

	taskCursor, err := r.coll.Aggregate(context.TODO(), taskPipeline)
	if err != nil {
		return nil, err
	}
	defer func(taskCursor *mongo.Cursor, ctx context.Context) {
		err := taskCursor.Close(ctx)
		if err != nil {
			return
		}
	}(taskCursor, context.TODO())

	var contributorIDs []primitive.ObjectID
	for taskCursor.Next(context.TODO()) {
		var result struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		if err := taskCursor.Decode(&result); err != nil {
			return nil, err
		}
		contributorIDs = append(contributorIDs, result.ID)
	}

	if len(contributorIDs) == 0 {
		return &countUser, nil
	}

	userPipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"_id", bson.D{{"$in", contributorIDs}}},
			{"is_deleted", false},
		}}},
		{{"$group", bson.D{
			{"_id", "$status"},
			{"count", bson.D{{"$sum", 1}}},
		}}},
	}

	userCursor, err := userRepo.coll.Aggregate(context.TODO(), userPipeline)
	if err != nil {
		return nil, err
	}
	defer func(userCursor *mongo.Cursor, ctx context.Context) {
		err := userCursor.Close(ctx)
		if err != nil {
			return
		}
	}(userCursor, context.TODO())

	for userCursor.Next(context.TODO()) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := userCursor.Decode(&result); err != nil {
			return nil, err
		}
		switch result.ID {
		case "active":
			countUser.Current.Active = result.Count
		case "not_active":
			countUser.Current.NotActive = result.Count
		}
	}
	countUser.Current.Total = countUser.Current.Active + countUser.Current.NotActive
	return &countUser, nil
}

func (r *TaskCollRepository) DeleteOneByID(_id primitive.ObjectID) error {
	filter := bson.M{
		"_id": _id,
	}
	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
		},
	}

	_, err := r.coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

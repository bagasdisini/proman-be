package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	_const "proman-backend/pkg/const"
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
	Current  CountTaskDetail `json:"current"`
	LastYear CountTaskDetail `json:"last_year"`
}

type CountUserActive struct {
	Total       int `json:"total"`
	HaveTask    int `json:"have_task"`
	NotHaveTask int `json:"not_have_task"`
}

type CountUser struct {
	Current  CountUserActive `json:"current"`
	LastYear CountUserActive `json:"last_year"`
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

func (r *TaskCollRepository) CreateOne(task *Task) error {
	_, err := r.coll.InsertOne(context.TODO(), task)
	if err != nil {
		return err
	}
	return nil
}

func (r *TaskCollRepository) CountTask(start, end time.Time) (*[]CountTaskDetail, error) {
	var count []CountTaskDetail
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
			{_const.TaskActive, bson.D{{"$sum", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$status", _const.TaskActive}}}, 1, 0}}}}}},
			{_const.TaskTesting, bson.D{{"$sum", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$status", _const.TaskTesting}}}, 1, 0}}}}}},
			{_const.TaskCompleted, bson.D{{"$sum", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$status", _const.TaskCompleted}}}, 1, 0}}}}}},
			{_const.TaskCancelled, bson.D{{"$sum", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$status", _const.TaskCancelled}}}, 1, 0}}}}}},
		}},
	}

	cursor, err := r.coll.Aggregate(context.Background(), mongo.Pipeline{match, group})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.TODO(), &count); err != nil {
		return nil, err
	}
	return &count, nil
}

func (r *TaskCollRepository) CountUserThatHaveTask(userRepo *UserCollRepository, start, end time.Time) (*CountUserActive, error) {
	result := &CountUserActive{}

	taskPipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"status", bson.M{"$in": bson.A{_const.TaskActive, _const.TaskTesting}}},
			{"start_date", bson.M{"$gt": start}},
			{"end_date", bson.M{"$lt": end}},
		}}},
		{{"$group", bson.D{
			{"_id", "$contributor"},
		}}},
	}

	cursor, err := r.coll.Aggregate(context.TODO(), taskPipeline)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			return
		}
	}(cursor, context.TODO())

	var contributorsWithTasks []primitive.ObjectID
	for cursor.Next(context.TODO()) {
		var result struct {
			ID []primitive.ObjectID `bson:"_id"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		contributorsWithTasks = append(contributorsWithTasks, result.ID...)
	}

	totalUsersCount, err := userRepo.coll.CountDocuments(context.TODO(), bson.M{"is_deleted": bson.M{"$ne": true}})
	if err != nil {
		return nil, err
	}

	haveTaskCount := len(contributorsWithTasks)
	notHaveTaskCount := totalUsersCount - int64(haveTaskCount)
	result.Total = int(totalUsersCount)
	result.HaveTask = haveTaskCount
	result.NotHaveTask = int(notHaveTaskCount)
	return result, nil
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

func (r *TaskCollRepository) DeleteAllByProjectID(projectID primitive.ObjectID) error {
	filter := bson.M{
		"project_id": projectID,
	}
	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
		},
	}

	_, err := r.coll.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

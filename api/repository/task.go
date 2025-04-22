package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"proman-backend/internal/pkg/const"
	"proman-backend/internal/pkg/util"
	"time"
)

type Task struct {
	ID          bson.ObjectID   `json:"_id" bson:"_id"`
	Name        string          `json:"name" bson:"name"`
	Description string          `json:"description" bson:"description"`
	StartDate   time.Time       `json:"start_date" bson:"start_date"`
	EndDate     time.Time       `json:"end_date" bson:"end_date"`
	Contributor []bson.ObjectID `json:"contributor" bson:"contributor"`
	Status      string          `json:"status" bson:"status"` // active, testing, completed, cancelled
	ProjectID   bson.ObjectID   `json:"project_id" bson:"project_id"`
	CreatedAt   time.Time       `json:"created_at" bson:"created_at"`
	IsDeleted   bool            `json:"-" bson:"is_deleted"`
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

type CountUserActive struct {
	Total       int `json:"total"`
	HaveTask    int `json:"have_task"`
	NotHaveTask int `json:"not_have_task"`
}

type TaskOverview struct {
	Start string `json:"start"`
	End   string `json:"end"`
	Count int    `json:"count"`
}

type TaskCollRepository struct {
	coll *mongo.Collection
}

func NewTaskCollRepository(db *mongo.Database) *TaskCollRepository {
	return &TaskCollRepository{
		coll: db.Collection("tasks"),
	}
}

func (r *TaskCollRepository) FindAll(cq *util.CommonQuery) ([]Task, error) {
	tasks := []Task{}
	filter := bson.M{"is_deleted": bson.M{"$ne": true}}

	if len(cq.Q) > 0 {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": bson.Regex{Pattern: cq.Q, Options: "i"}}},
			{"description": bson.M{"$regex": bson.Regex{Pattern: cq.Q, Options: "i"}}},
		}
	}

	if len(cq.Status) > 0 && _const.IsValidTaskStatus(cq.Status) {
		filter["status"] = cq.Status
	}

	if cq.UserId != bson.NilObjectID {
		filter["contributor"] = cq.UserId
	}

	if cq.ProjectId != bson.NilObjectID {
		filter["project_id"] = cq.ProjectId
	}

	if existingOr, ok := filter["$or"]; ok {
		filter["$or"] = append(existingOr.([]bson.M), bson.M{
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
		filter["$or"] = []bson.M{
			{
				"start_date": bson.M{"$lt": cq.End},
				"end_date":   bson.M{"$gte": cq.Start},
			},
			{
				"start_date": bson.M{"$gte": cq.Start, "$lt": cq.End},
			},
		}
	}

	cursor, err := r.coll.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	if err := cursor.All(context.TODO(), &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *TaskCollRepository) FindOneByID(_id bson.ObjectID) (*Task, error) {
	user := Task{}
	filter := bson.M{
		"_id":        _id,
		"is_deleted": bson.M{"$ne": true},
	}

	err := r.coll.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *TaskCollRepository) CreateOne(task *Task) error {
	_, err := r.coll.InsertOne(context.TODO(), task)
	if err != nil {
		return err
	}
	return nil
}

func (r *TaskCollRepository) CountTask(cq *util.CommonQuery) ([]CountTaskDetail, error) {
	count := []CountTaskDetail{}

	matchStage := bson.D{{"is_deleted", bson.M{"$ne": true}}}

	if len(cq.Q) > 0 {
		matchStage = append(matchStage, bson.E{Key: "$or", Value: bson.A{
			bson.D{{"name", bson.Regex{Pattern: cq.Q, Options: "i"}}},
			bson.D{{"description", bson.Regex{Pattern: cq.Q, Options: "i"}}},
		}})
	}

	if len(cq.Status) > 0 && _const.IsValidTaskStatus(cq.Status) {
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
	return count, nil
}

func (r *TaskCollRepository) CountUserTask(userRepo *UserCollRepository) (*CountUserActive, error) {
	result := &CountUserActive{}

	taskPipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"status", bson.M{"$in": bson.A{_const.TaskActive, _const.TaskTesting}}},
		}}},
		{{"$group", bson.D{
			{"_id", "$contributor"},
		}}},
	}

	cursor, err := r.coll.Aggregate(context.TODO(), taskPipeline)
	if err != nil {
		return nil, err
	}

	contributorsWithTasks := []bson.ObjectID{}
	for cursor.Next(context.TODO()) {
		result := struct {
			ID []bson.ObjectID `bson:"_id"`
		}{}
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

func (r *TaskCollRepository) UpdateOneByID(task *Task) error {
	filter := bson.M{
		"_id":        task.ID,
		"is_deleted": bson.M{"$ne": true},
	}
	update := bson.M{"$set": task}

	_, err := r.coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (r *TaskCollRepository) DeleteOneByID(_id bson.ObjectID) error {
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

func (r *TaskCollRepository) DeleteAllByProjectID(projectID bson.ObjectID) error {
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

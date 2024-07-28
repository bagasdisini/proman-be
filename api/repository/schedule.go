package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"proman-backend/pkg/util"
	"time"
)

type Schedule struct {
	ID          primitive.ObjectID   `json:"_id" bson:"_id"`
	Name        string               `json:"name" bson:"name"`
	Description string               `json:"description" bson:"description"`
	StartDate   time.Time            `json:"start_date" bson:"start_date"`
	EndDate     time.Time            `json:"end_date" bson:"end_date"`
	Contributor []primitive.ObjectID `json:"contributor" bson:"contributor"`
	Type        string               `json:"type" bson:"type"` // meeting, discussion, review, presentation
	CreatedAt   time.Time            `json:"created_at" bson:"created_at"`
	IsDeleted   bool                 `json:"-" bson:"is_deleted"`
}

type ScheduleCollRepository struct {
	coll *mongo.Collection
}

func NewScheduleRepository(db *mongo.Database) *ScheduleCollRepository {
	return &ScheduleCollRepository{
		coll: db.Collection("schedules"),
	}
}

func (r *ScheduleCollRepository) FindAll(cq *util.ScheduleQuery) ([]Schedule, error) {
	var schedules []Schedule
	filter := bson.M{"is_deleted": false}

	if len(cq.Q) > 0 {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": primitive.Regex{Pattern: cq.Q, Options: "i"}}},
			{"description": bson.M{"$regex": primitive.Regex{Pattern: cq.Q, Options: "i"}}},
		}
	}

	if len(cq.Type) > 0 {
		filter["type"] = cq.Type
	}

	if len(cq.Contributor) > 0 {
		filter["contributor"] = bson.M{"$in": cq.Contributor}
	}

	if !cq.Start.IsZero() {
		filter["start_date"] = bson.M{"$gte": cq.Start}
	}

	if !cq.End.IsZero() {
		filter["end_date"] = bson.M{"$lte": cq.End}
	}

	cursor, err := r.coll.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	if err := cursor.All(context.TODO(), &schedules); err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *ScheduleCollRepository) FindAllByDateRange(startDate, endDate time.Time) ([]Schedule, error) {
	filter := bson.M{
		"start_date": bson.M{"$gte": startDate},
		"end_date":   bson.M{"$lte": endDate},
		"is_deleted": false,
	}

	cursor, err := r.coll.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	var schedules []Schedule
	if err := cursor.All(context.TODO(), &schedules); err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *ScheduleCollRepository) CreateOne(schedule *Schedule) error {
	_, err := r.coll.InsertOne(context.TODO(), schedule)
	if err != nil {
		return err
	}
	return nil
}

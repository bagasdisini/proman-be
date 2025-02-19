package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"proman-backend/internal/pkg/const"
	"proman-backend/internal/pkg/util"
	"time"
)

type Schedule struct {
	ID          bson.ObjectID   `json:"_id" bson:"_id"`
	Name        string          `json:"name" bson:"name"`
	Description string          `json:"description" bson:"description"`
	StartDate   time.Time       `json:"start_date" bson:"start_date"`
	EndDate     time.Time       `json:"end_date" bson:"end_date"`
	StartTime   string          `json:"start_time" bson:"start_time"` // 24-hour format (HH:MM)
	EndTime     string          `json:"end_time" bson:"end_time"`     // 24-hour format (HH:MM)
	Contributor []bson.ObjectID `json:"contributor" bson:"contributor"`
	Type        string          `json:"type" bson:"type"` // meeting, discussion, review, presentation
	CreatedAt   time.Time       `json:"created_at" bson:"created_at"`
	IsDeleted   bool            `json:"-" bson:"is_deleted"`
}

type ScheduleCollRepository struct {
	coll *mongo.Collection
}

func NewScheduleCollRepository(db *mongo.Database) *ScheduleCollRepository {
	return &ScheduleCollRepository{
		coll: db.Collection("schedules"),
	}
}

func (r *ScheduleCollRepository) FindAll(cq *util.CommonQuery) ([]Schedule, error) {
	schedules := []Schedule{}
	filter := bson.M{"is_deleted": false}

	if len(cq.Q) > 0 {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": bson.Regex{Pattern: cq.Q, Options: "i"}}},
			{"description": bson.M{"$regex": bson.Regex{Pattern: cq.Q, Options: "i"}}},
		}
	}

	if len(cq.Type) > 0 && _const.IsValidScheduleType(cq.Type) {
		filter["type"] = cq.Type
	}

	if cq.UserId != bson.NilObjectID {
		filter["contributor"] = cq.UserId
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

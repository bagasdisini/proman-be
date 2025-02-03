package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"time"
)

type Code struct {
	ID        bson.ObjectID `json:"_id" bson:"_id"`
	UserID    bson.ObjectID `json:"user_id" bson:"user_id"`
	Email     string        `json:"email" bson:"email"`
	Code      string        `json:"code" bson:"code"`
	IsUsed    bool          `json:"is_used" bson:"is_used"`
	ExpiredAt time.Time     `json:"expired_at" bson:"expired_at"`
	CreatedAt time.Time     `json:"created_at" bson:"created_at"`
}

type CodeCollRepository struct {
	coll *mongo.Collection
}

func NewCodeCollRepository(db *mongo.Database) *CodeCollRepository {
	return &CodeCollRepository{
		coll: db.Collection("codes"),
	}
}

func (r *CodeCollRepository) FindOneByCode(code string) (*Code, error) {
	doc := Code{}
	filter := bson.M{
		"code": code,
	}

	err := r.coll.FindOne(context.TODO(), filter).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *CodeCollRepository) FindActiveOneByCode(code string) (*Code, error) {
	doc := Code{}
	filter := bson.M{
		"code":       code,
		"is_used":    false,
		"expired_at": bson.M{"$gte": time.Now()},
	}

	err := r.coll.FindOne(context.TODO(), filter).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *CodeCollRepository) FindActiveOneByUserID(userID bson.ObjectID) (*Code, error) {
	doc := Code{}
	filter := bson.M{
		"user_id":    userID,
		"is_used":    false,
		"expired_at": bson.M{"$gte": time.Now()},
	}

	err := r.coll.FindOne(context.TODO(), filter).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *CodeCollRepository) InsertOne(code *Code) (*Code, error) {
	doc := Code{}
	dataInsert, err := r.coll.InsertOne(context.TODO(), code)
	if err != nil {
		return nil, err
	}

	insertedID := dataInsert.InsertedID.(bson.ObjectID)
	err = r.coll.FindOne(context.TODO(), bson.M{"_id": insertedID}).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *CodeCollRepository) UpdateOne(code *Code) (*Code, error) {
	doc := Code{}
	filter := bson.M{"_id": code.ID}
	update := bson.M{"$set": code}

	err := r.coll.FindOneAndUpdate(context.TODO(), filter, update).Decode(&doc)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

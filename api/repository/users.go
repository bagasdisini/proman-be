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

type User struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	Email     string             `json:"email" bson:"email"`
	Password  string             `json:"password" bson:"password"`
	Name      string             `json:"name" bson:"name"`
	Role      string             `json:"role" bson:"role"`
	Position  string             `json:"position" bson:"position"`
	Avatar    string             `json:"avatar" bson:"avatar"`
	Phone     string             `json:"phone" bson:"phone"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	IsDeleted bool               `json:"-" bson:"is_deleted"`
}

func (u *User) MarshalJSON() ([]byte, error) {
	type Alias User
	var url string
	if u.Avatar != "" {
		url = "https://" + config.S3.Bucket
		if !strings.Contains(config.S3.EndPoint, "https://") {
			url = url + "." + config.S3.EndPoint + "/" + config.AWS.AvatarDir + "/" + u.Avatar
		} else {
			url = url + "." + config.S3.EndPoint[8:] + "/" + config.AWS.AvatarDir + "/" + u.Avatar
		}
	}
	return json.Marshal(&struct {
		*Alias
		Avatar string `json:"avatar" bson:"avatar"`
	}{
		Alias:  (*Alias)(u),
		Avatar: url,
	})
}

type UserCollRepository struct {
	coll *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserCollRepository {
	return &UserCollRepository{
		coll: db.Collection("users"),
	}
}

func (r *UserCollRepository) FindOneByID(_id primitive.ObjectID) (*User, error) {
	var user *User
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

func (r *UserCollRepository) FindOneByEmail(email string) (*User, error) {
	var user *User
	filter := bson.M{
		"email":      email,
		"is_deleted": bson.M{"$ne": true},
	}

	err := r.coll.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserCollRepository) Insert(userData *User) (*User, error) {
	var data *User
	dataInsert, err := r.coll.InsertOne(context.TODO(), userData)
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

func (r *UserCollRepository) Update(userData *User) (*User, error) {
	var data *User
	filter := bson.M{"_id": userData.ID, "is_deleted": bson.M{"$ne": true}}
	update := bson.M{"$set": userData}

	err := r.coll.FindOneAndUpdate(context.TODO(), filter, update).Decode(&data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *UserCollRepository) Check() bool {
	filter := bson.M{"is_deleted": bson.M{"$ne": true}}
	count, err := r.coll.CountDocuments(context.TODO(), filter)
	if err != nil {
		return false
	}
	if count > 0 {
		return true
	}
	return false
}

package repository

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"proman-backend/config"
	"proman-backend/internal/pkg/util"
	"strings"
	"time"
)

type User struct {
	ID        bson.ObjectID `json:"_id" bson:"_id"`
	Email     string        `json:"email" bson:"email"`
	Password  string        `json:"-" bson:"password"`
	Name      string        `json:"name" bson:"name"`
	Position  string        `json:"position" bson:"position"`
	Avatar    string        `json:"avatar" bson:"avatar"`
	Phone     string        `json:"phone" bson:"phone"`
	CreatedAt time.Time     `json:"created_at" bson:"created_at"`
	IsDeleted bool          `json:"-" bson:"is_deleted"`
}

func (u *User) MarshalJSON() ([]byte, error) {
	type Alias User
	url := ""
	if u.Avatar != "" {
		url = "https://" + config.S3.Bucket
		if !strings.Contains(config.S3.EndPoint, "https://") {
			url = url + "." + config.S3.EndPoint + "/" + u.Avatar
		} else {
			url = url + "." + config.S3.EndPoint[8:] + "/" + u.Avatar
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

func NewUserCollRepository(db *mongo.Database) *UserCollRepository {
	return &UserCollRepository{
		coll: db.Collection("users"),
	}
}

func (r *UserCollRepository) FindAllUsers(cq *util.CommonQuery) ([]map[string]interface{}, error) {
	users := []map[string]interface{}{}

	matchStage := bson.D{{"is_deleted", bson.D{{"$ne", true}}}}

	if len(cq.Q) > 0 {
		matchStage = append(matchStage, bson.E{Key: "$or", Value: bson.A{
			bson.D{{"email", bson.Regex{Pattern: cq.Q, Options: "i"}}},
			bson.D{{"name", bson.Regex{Pattern: cq.Q, Options: "i"}}},
		}})
	}

	skip := (cq.Page - 1) * cq.Limit
	limit := cq.Limit

	pipeline := mongo.Pipeline{
		{{"$match", matchStage}},
		{{"$lookup", bson.D{
			{"from", "projects"},
			{"let", bson.D{{"user_id", "$_id"}}},
			{"pipeline", mongo.Pipeline{
				{{"$match", bson.D{
					{"$expr", bson.D{
						{"$and", bson.A{
							bson.D{{"$eq", bson.A{"$is_deleted", false}}},
							bson.D{{"$in", bson.A{"$$user_id", "$contributor"}}},
						}},
					}},
				}}},
				{{"$count", "total"}},
			}},
			{"as", "projects"},
		}}},
		{{"$addFields", bson.D{
			{"total_project", bson.D{
				{"$cond", bson.D{
					{"if", bson.D{{"$gt", bson.A{bson.D{{"$size", "$projects"}}, 0}}}},
					{"then", bson.D{{"$arrayElemAt", bson.A{"$projects.total", 0}}}},
					{"else", 0},
				}},
			}},
		}}},
		{{"$lookup", bson.D{
			{"from", "tasks"},
			{"let", bson.D{{"user_id", "$_id"}}},
			{"pipeline", mongo.Pipeline{
				{{"$match", bson.D{
					{"$expr", bson.D{
						{"$and", bson.A{
							bson.D{{"$eq", bson.A{"$is_deleted", false}}},
							bson.D{{"$in", bson.A{"$$user_id", "$contributor"}}},
						}},
					}},
				}}},
				{{"$count", "total"}},
			}},
			{"as", "tasks"},
		}}},
		{{"$addFields", bson.D{
			{"total_task", bson.D{
				{"$cond", bson.D{
					{"if", bson.D{{"$gt", bson.A{bson.D{{"$size", "$tasks"}}, 0}}}},
					{"then", bson.D{{"$arrayElemAt", bson.A{"$tasks.total", 0}}}},
					{"else", 0},
				}},
			}},
		}}},
		{{"$unset", bson.A{"projects", "tasks"}}},
		{{"$sort", bson.D{{"created_at", cq.Sort}}}},
		{{"$skip", skip}},
		{{"$limit", limit}},
	}

	cursor, err := r.coll.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}

	for cursor.Next(context.TODO()) {
		user := map[string]interface{}{}
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserCollRepository) FindOneByID(_id bson.ObjectID) (*User, error) {
	user := User{}
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

func (r *UserCollRepository) FindOneByEmail(email string) (*User, error) {
	user := User{}
	filter := bson.M{
		"email":      email,
		"is_deleted": bson.M{"$ne": true},
	}

	err := r.coll.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserCollRepository) Insert(userData *User) (*User, error) {
	data := User{}
	dataInsert, err := r.coll.InsertOne(context.TODO(), userData)
	if err != nil {
		return nil, err
	}

	insertedID := dataInsert.InsertedID.(bson.ObjectID)
	err = r.coll.FindOne(context.TODO(), bson.M{"_id": insertedID}).Decode(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *UserCollRepository) Update(userData *User) (*User, error) {
	data := User{}
	filter := bson.M{"_id": userData.ID, "is_deleted": bson.M{"$ne": true}}
	update := bson.M{"$set": userData}

	err := r.coll.FindOneAndUpdate(context.TODO(), filter, update).Decode(&data)
	if err != nil {
		return nil, err
	}
	return userData, nil
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

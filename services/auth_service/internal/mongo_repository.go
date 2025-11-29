package auth

import (
	"context"

	authpb "github.com/Miguel-Pezzini/GoMessenger/services/auth_service/internal/pb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{
		collection: db.Collection("users"),
	}
}

func (r *MongoRepository) Create(ctx context.Context, registerUserRequest *authpb.RegisterRequest) (*User, error) {
	userMongo := UserMongo{
		Username: registerUserRequest.Username,
		Password: registerUserRequest.Password,
	}

	result, err := r.collection.InsertOne(ctx, userMongo)
	if err != nil {
		return nil, err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		userMongo.ID = oid
	}

	user := &User{
		ID:       userMongo.ID.Hex(),
		Username: userMongo.Username,
		Password: userMongo.Password,
	}

	return user, nil
}

func (r *MongoRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	var userMongo UserMongo
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&userMongo)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:       userMongo.ID.Hex(),
		Username: userMongo.Username,
		Password: userMongo.Password,
	}

	return user, nil
}

func (r *MongoRepository) SearchByUsername(ctx context.Context, query string, limit int) ([]*User, error) {
	// Use regex for case-insensitive partial match
	filter := bson.M{
		"username": bson.M{"$regex": query, "$options": "i"},
	}

	opts := options.Find().SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var usersMongo []UserMongo
	if err := cursor.All(ctx, &usersMongo); err != nil {
		return nil, err
	}

	users := make([]*User, len(usersMongo))
	for i, um := range usersMongo {
		users[i] = &User{
			ID:       um.ID.Hex(),
			Username: um.Username,
			Password: um.Password,
		}
	}

	return users, nil
}

func (r *MongoRepository) GetUsersByIDs(ctx context.Context, ids []string) ([]*User, error) {
	objectIDs := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		if oid, err := primitive.ObjectIDFromHex(id); err == nil {
			objectIDs = append(objectIDs, oid)
		}
	}

	if len(objectIDs) == 0 {
		return []*User{}, nil
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var usersMongo []UserMongo
	if err := cursor.All(ctx, &usersMongo); err != nil {
		return nil, err
	}

	users := make([]*User, len(usersMongo))
	for i, um := range usersMongo {
		users[i] = &User{
			ID:       um.ID.Hex(),
			Username: um.Username,
			Password: um.Password,
		}
	}

	return users, nil
}

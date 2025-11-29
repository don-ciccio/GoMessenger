package chat

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{
		collection: db.Collection("messages"),
	}
}

func (r *MongoRepository) Create(ctx context.Context, message *MessageDB) (*MessageDB, error) {
	result, err := r.collection.InsertOne(ctx, message)
	if err != nil {
		return nil, err
	}

	oid, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("failed to convert inserted ID")
	}
	message.Id = oid.Hex()
	return message, nil
}

func (r *MongoRepository) FindByConversationID(ctx context.Context, conversationID string, limit, offset int) ([]*MessageDB, error) {
	filter := map[string]interface{}{
		"conversation_id": conversationID,
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(map[string]int{"timestamp": 1})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []*MessageDB
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

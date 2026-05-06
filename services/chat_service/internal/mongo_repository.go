package chat

import (
	"context"
	"fmt"

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
		SetSort(map[string]int{"timestamp": -1})

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

// UpdateViewedStatus updates a single message's status with a monotonic rank guard.
// Only upgrades (sent→delivered→seen) are allowed; downgrades are silently ignored.
func (r *MongoRepository) UpdateViewedStatus(ctx context.Context, messageID string, status string) error {
	objectID, err := primitive.ObjectIDFromHex(messageID)
	if err != nil {
		return err
	}

	// Fetch current status
	var current MessageDB
	if err := r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&current); err != nil {
		return err
	}

	// Only upgrade, never downgrade
	if ViewedStatusRank(NormalizeViewedStatus(current.ViewedStatus)) >= ViewedStatusRank(status) {
		return nil
	}

	_, err = r.collection.UpdateOne(ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{"viewed_status": status}},
	)
	return err
}

// MarkConversationSeen batch-updates all messages from a given sender in a conversation
// that aren't already "seen". Used when the recipient opens the chat.
func (r *MongoRepository) MarkConversationSeen(ctx context.Context, conversationID string, senderID string) error {
	filter := bson.M{
		"conversation_id": conversationID,
		"sender_id":       senderID,
		"viewed_status":   bson.M{"$ne": ViewedStatusSeen},
	}
	update := bson.M{
		"$set": bson.M{"viewed_status": ViewedStatusSeen},
	}
	_, err := r.collection.UpdateMany(ctx, filter, update)
	return err
}

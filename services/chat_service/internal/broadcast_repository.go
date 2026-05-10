package chat

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BroadcastRepository interface {
	Create(ctx context.Context, broadcast *Broadcast) (*Broadcast, error)
	FindByID(ctx context.Context, id string) (*Broadcast, error)
	ListBySender(ctx context.Context, senderID string, limit, offset int) ([]*Broadcast, error)
	UpdateStatus(ctx context.Context, id, status string) error
	SetCompleted(ctx context.Context, id string, successCount, failureCount int) error
	AddConversationID(ctx context.Context, broadcastID, conversationID string) error
	GetLastBroadcastTime(ctx context.Context, senderID string) (time.Time, error)
}

type MongoBroadcastRepository struct {
	collection *mongo.Collection
}

func NewMongoBroadcastRepository(db *mongo.Database) *MongoBroadcastRepository {
	return &MongoBroadcastRepository{
		collection: db.Collection("broadcasts"),
	}
}

func (r *MongoBroadcastRepository) Create(ctx context.Context, broadcast *Broadcast) (*Broadcast, error) {
	result, err := r.collection.InsertOne(ctx, broadcast)
	if err != nil {
		return nil, err
	}
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		broadcast.ID = oid.Hex()
	}
	return broadcast, nil
}

func (r *MongoBroadcastRepository) FindByID(ctx context.Context, id string) (*Broadcast, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var broadcast Broadcast
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&broadcast)
	if err != nil {
		return nil, err
	}
	return &broadcast, nil
}

func (r *MongoBroadcastRepository) ListBySender(ctx context.Context, senderID string, limit, offset int) ([]*Broadcast, error) {
	filter := bson.M{"sender_id": senderID}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var broadcasts []*Broadcast
	if err := cursor.All(ctx, &broadcasts); err != nil {
		return nil, err
	}
	return broadcasts, nil
}

func (r *MongoBroadcastRepository) UpdateStatus(ctx context.Context, id, status string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid},
		bson.M{"$set": bson.M{"status": status}})
	return err
}

func (r *MongoBroadcastRepository) SetCompleted(ctx context.Context, id string, successCount, failureCount int) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid},
		bson.M{"$set": bson.M{
			"status":        "completed",
			"success_count": successCount,
			"failure_count": failureCount,
			"completed_at":  time.Now(),
		}})
	return err
}

func (r *MongoBroadcastRepository) AddConversationID(ctx context.Context, broadcastID, conversationID string) error {
	oid, err := primitive.ObjectIDFromHex(broadcastID)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid},
		bson.M{"$addToSet": bson.M{"conversation_ids": conversationID}})
	return err
}

func (r *MongoBroadcastRepository) GetLastBroadcastTime(ctx context.Context, senderID string) (time.Time, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})
	var broadcast Broadcast
	err := r.collection.FindOne(ctx, bson.M{"sender_id": senderID}, opts).Decode(&broadcast)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return time.Time{}, nil // No previous broadcast
		}
		return time.Time{}, err
	}
	return broadcast.CreatedAt, nil
}

package chat

import (
	"context"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ConversationRepository interface {
	Create(ctx context.Context, participants []string) (*Conversation, error)
	FindByParticipants(ctx context.Context, participants []string) (*Conversation, error)
	FindByID(ctx context.Context, id string) (*Conversation, error)
	ListByUserID(ctx context.Context, userID string) ([]*Conversation, error)
	UpdateLastMessage(ctx context.Context, conversationID string, message string) error
}

type MongoConversationRepository struct {
	db *mongo.Database
}

func NewMongoConversationRepository(db *mongo.Database) *MongoConversationRepository {
	return &MongoConversationRepository{db: db}
}

func (r *MongoConversationRepository) Create(ctx context.Context, participants []string) (*Conversation, error) {
	// Sort participants to ensure consistent lookup
	sort.Strings(participants)

	conversation := &Conversation{
		Participants:  participants,
		LastMessage:   "",
		LastMessageAt: time.Now(),
		CreatedAt:     time.Now(),
	}

	result, err := r.db.Collection("conversations").InsertOne(ctx, conversation)
	if err != nil {
		return nil, err
	}

	conversation.ID = result.InsertedID.(primitive.ObjectID).Hex()
	return conversation, nil
}

func (r *MongoConversationRepository) FindByParticipants(ctx context.Context, participants []string) (*Conversation, error) {
	sort.Strings(participants)

	filter := bson.M{"participants": bson.M{"$all": participants, "$size": len(participants)}}

	var conversation Conversation
	err := r.db.Collection("conversations").FindOne(ctx, filter).Decode(&conversation)
	if err != nil {
		return nil, err
	}

	return &conversation, nil
}

func (r *MongoConversationRepository) FindByID(ctx context.Context, id string) (*Conversation, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var conversation Conversation
	err = r.db.Collection("conversations").FindOne(ctx, bson.M{"_id": objectID}).Decode(&conversation)
	if err != nil {
		return nil, err
	}

	return &conversation, nil
}

func (r *MongoConversationRepository) ListByUserID(ctx context.Context, userID string) ([]*Conversation, error) {
	filter := bson.M{"participants": userID}
	opts := options.Find().SetSort(bson.D{{Key: "last_message_at", Value: -1}})

	cursor, err := r.db.Collection("conversations").Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var conversations []*Conversation
	if err := cursor.All(ctx, &conversations); err != nil {
		return nil, err
	}

	return conversations, nil
}

func (r *MongoConversationRepository) UpdateLastMessage(ctx context.Context, conversationID string, message string) error {
	objectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"last_message":    message,
			"last_message_at": time.Now(),
		},
	}

	_, err = r.db.Collection("conversations").UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

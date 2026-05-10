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
	Create(ctx context.Context, participants []string, shopID string) (*Conversation, error)
	FindByParticipants(ctx context.Context, participants []string) (*Conversation, error)
	FindByID(ctx context.Context, id string) (*Conversation, error)
	ListByUserID(ctx context.Context, userID string, shopID string) ([]*Conversation, error)
	ListArchivedByUserID(ctx context.Context, userID string) ([]*Conversation, error)
	UpdateLastMessage(ctx context.Context, conversationID string, message string, senderID string) error
	ArchiveForUser(ctx context.Context, conversationID string, userID string) error
	UnarchiveForUser(ctx context.Context, conversationID string, userID string) error
}

type MongoConversationRepository struct {
	db *mongo.Database
}

func NewMongoConversationRepository(db *mongo.Database) *MongoConversationRepository {
	return &MongoConversationRepository{db: db}
}

func (r *MongoConversationRepository) Create(ctx context.Context, participants []string, shopID string) (*Conversation, error) {
	// Sort participants to ensure consistent lookup
	sort.Strings(participants)

	conversation := &Conversation{
		Participants:  participants,
		ShopID:        shopID,
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

func (r *MongoConversationRepository) ListByUserID(ctx context.Context, userID string, shopID string) ([]*Conversation, error) {
	filter := bson.M{
		"participants": userID,
		"archived_by":  bson.M{"$nin": bson.A{userID}},
	}
	if shopID != "" {
		filter["shop_id"] = shopID
	}
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

func (r *MongoConversationRepository) ListArchivedByUserID(ctx context.Context, userID string) ([]*Conversation, error) {
	filter := bson.M{
		"participants": userID,
		"archived_by":  userID,
	}
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

func (r *MongoConversationRepository) ArchiveForUser(ctx context.Context, conversationID string, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return err
	}
	_, err = r.db.Collection("conversations").UpdateOne(ctx,
		bson.M{"_id": objectID},
		bson.M{"$addToSet": bson.M{"archived_by": userID}},
	)
	return err
}

func (r *MongoConversationRepository) UnarchiveForUser(ctx context.Context, conversationID string, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return err
	}
	_, err = r.db.Collection("conversations").UpdateOne(ctx,
		bson.M{"_id": objectID},
		bson.M{"$pull": bson.M{"archived_by": userID}},
	)
	return err
}

func (r *MongoConversationRepository) UpdateLastMessage(ctx context.Context, conversationID string, message string, senderID string) error {
	objectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		return err
	}

	setFields := bson.M{
		"last_message":    message,
		"last_message_at": time.Now(),
	}
	if senderID != "" {
		setFields["last_message_sender_id"] = senderID
	}

	update := bson.M{
		"$set":   setFields,
		"$unset": bson.M{"archived_by": ""},
	}

	_, err = r.db.Collection("conversations").UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

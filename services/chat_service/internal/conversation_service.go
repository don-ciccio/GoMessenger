package chat

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

type ConversationService struct {
	repo ConversationRepository
}

func NewConversationService(repo ConversationRepository) *ConversationService {
	return &ConversationService{repo: repo}
}

// GetOrCreateConversation finds an existing conversation or creates a new one
func (s *ConversationService) GetOrCreateConversation(ctx context.Context, participants []string) (*Conversation, error) {
	// Try to find existing conversation
	conversation, err := s.repo.FindByParticipants(ctx, participants)
	if err == nil {
		return conversation, nil
	}

	// If not found, create new
	if err == mongo.ErrNoDocuments {
		return s.repo.Create(ctx, participants)
	}

	return nil, err
}

func (s *ConversationService) GetConversation(ctx context.Context, id string) (*Conversation, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ConversationService) ListUserConversations(ctx context.Context, userID string) ([]*Conversation, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *ConversationService) UpdateLastMessage(ctx context.Context, conversationID string, message string) error {
	return s.repo.UpdateLastMessage(ctx, conversationID, message)
}

// GetConversationMessages retrieves messages for a conversation
func (s *ConversationService) GetConversationMessages(ctx context.Context, repo Repository, conversationID string, limit, offset int) ([]*MessageDB, error) {
	messages, err := repo.FindByConversationID(ctx, conversationID, limit, offset)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// ValidateUserInConversation checks if a user is a participant in the conversation
func (s *ConversationService) ValidateUserInConversation(ctx context.Context, conversationID, userID string) error {
	conversation, err := s.repo.FindByID(ctx, conversationID)
	if err != nil {
		return err
	}

	for _, participant := range conversation.Participants {
		if participant == userID {
			return nil
		}
	}

	return errors.New("user is not a participant in this conversation")
}

package chat

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, messageDB *MessageDB) (*MessageDB, error)
	FindByConversationID(ctx context.Context, conversationID string, limit, offset int) ([]*MessageDB, error)
	UpdateViewedStatus(ctx context.Context, messageID string, status string) error
	MarkConversationSeen(ctx context.Context, conversationID string, senderID string) error
}

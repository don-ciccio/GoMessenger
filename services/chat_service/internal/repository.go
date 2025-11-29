package chat

import (
	"context"
)

type Repository interface {
	Create(ctx context.Context, messageDB *MessageDB) (*MessageDB, error)
	FindByConversationID(ctx context.Context, conversationID string, limit, offset int) ([]*MessageDB, error)
}

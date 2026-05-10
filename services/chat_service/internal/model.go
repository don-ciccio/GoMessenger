package chat

import "time"

// Viewed-status lifecycle: sent → delivered → seen
const (
	ViewedStatusSent      = "sent"
	ViewedStatusDelivered = "delivered"
	ViewedStatusSeen      = "seen"
)

// Conversation represents a chat between 2+ users
type Conversation struct {
	ID            string    `bson:"_id,omitempty" json:"id"`
	Participants  []string  `bson:"participants" json:"participants"` // User IDs
	ShopID        string    `bson:"shop_id,omitempty" json:"shop_id,omitempty"` // Shopify store domain (for merchant conversations)
	LastMessage         string    `bson:"last_message" json:"last_message"`
	LastMessageAt       time.Time `bson:"last_message_at" json:"last_message_at"`
	LastMessageSenderID string    `bson:"last_message_sender_id,omitempty" json:"last_message_sender_id,omitempty"`
	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
	ArchivedBy    []string  `bson:"archived_by,omitempty" json:"archived_by,omitempty"`
}

type ConversationRequest struct {
	Participants []string `json:"participants"`
	ShopID       string   `json:"shop_id,omitempty"` // Shopify store domain (optional)
}

type MessageRequest struct {
	ConversationID string `json:"conversation_id"`
	SenderID       string `json:"sender_id"`
	ReceiverID     string `json:"receiver_id,omitempty"` // Deprecated, use ConversationID
	Content        string `json:"content"`
	Timestamp      int64  `json:"timestamp,omitempty"`
	BroadcastID    string `json:"broadcast_id,omitempty"`
	Tag            string `json:"tag,omitempty"` // Custom push notification title for broadcasts
}

type MessageDB struct {
	Id             string `bson:"_id,omitempty" json:"id"`
	ConversationID string `bson:"conversation_id" json:"conversation_id"`
	SenderID       string `bson:"sender_id" json:"sender_id"`
	ReceiverID     string `bson:"receiver_id,omitempty" json:"receiver_id,omitempty"` // Deprecated
	Content        string `bson:"content" json:"content"`
	Timestamp      int64  `bson:"timestamp" json:"timestamp"`
	ViewedStatus   string `bson:"viewed_status,omitempty" json:"viewed_status,omitempty"`
	BroadcastID    string `bson:"broadcast_id,omitempty" json:"broadcast_id,omitempty"`
	Tag            string `bson:"tag,omitempty" json:"tag,omitempty"`
}

type MessageResponse struct {
	Id             string   `json:"id"`
	ConversationID string   `json:"conversation_id"`
	SenderID       string   `json:"sender_id"`
	ReceiverID     string   `json:"receiver_id,omitempty"`
	Recipients     []string `json:"recipients,omitempty"` // List of user IDs to receive the message
	Content        string   `json:"content"`
	Timestamp      int64    `json:"timestamp,omitempty"`
	ViewedStatus   string   `json:"viewed_status,omitempty"`
	BroadcastID    string   `json:"broadcast_id,omitempty"`
	Tag            string   `json:"tag,omitempty"`
}

// InteractionEvent is received from the gateway via Redis Pub/Sub
// when a client sends message_delivered or message_seen.
type InteractionEvent struct {
	Type           string `json:"type"`
	ActorUserID    string `json:"actor_user_id"`
	TargetUserID   string `json:"target_user_id"`
	MessageID      string `json:"message_id,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
	ViewedStatus   string `json:"viewed_status,omitempty"`
}

// NormalizeViewedStatus maps any input to a known status; defaults to "sent".
func NormalizeViewedStatus(status string) string {
	switch status {
	case ViewedStatusDelivered:
		return ViewedStatusDelivered
	case ViewedStatusSeen:
		return ViewedStatusSeen
	default:
		return ViewedStatusSent
	}
}

// ViewedStatusRank returns a numeric rank so we can enforce monotonic upgrades.
func ViewedStatusRank(status string) int {
	switch NormalizeViewedStatus(status) {
	case ViewedStatusDelivered:
		return 1
	case ViewedStatusSeen:
		return 2
	default:
		return 0
	}
}

// Broadcast represents a mass message sent to multiple merchants
type Broadcast struct {
	ID              string    `bson:"_id,omitempty" json:"id"`
	SenderID        string    `bson:"sender_id" json:"sender_id"`
	Content         string    `bson:"content" json:"content"`
	Tag             string    `bson:"tag" json:"tag"`
	RecipientIDs    []string  `bson:"recipient_ids" json:"recipient_ids"`
	ConversationIDs []string  `bson:"conversation_ids,omitempty" json:"conversation_ids,omitempty"`
	TotalCount      int       `bson:"total_count" json:"total_count"`
	SuccessCount    int       `bson:"success_count" json:"success_count"`
	FailureCount    int       `bson:"failure_count" json:"failure_count"`
	Status          string    `bson:"status" json:"status"` // pending | sending | completed | failed
	CreatedAt       time.Time `bson:"created_at" json:"created_at"`
	CompletedAt     time.Time `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}

type BroadcastRequest struct {
	RecipientIDs []string `json:"recipient_ids"`
	Content      string   `json:"content"`
	Tag          string   `json:"tag"`
}

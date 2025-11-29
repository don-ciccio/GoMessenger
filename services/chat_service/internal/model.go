package chat

import "time"

// Conversation represents a chat between 2+ users
type Conversation struct {
	ID            string    `bson:"_id,omitempty" json:"id"`
	Participants  []string  `bson:"participants" json:"participants"` // User IDs
	LastMessage   string    `bson:"last_message" json:"last_message"`
	LastMessageAt time.Time `bson:"last_message_at" json:"last_message_at"`
	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
}

type ConversationRequest struct {
	Participants []string `json:"participants"`
}

type MessageRequest struct {
	ConversationID string `json:"conversation_id"`
	SenderID       string `json:"sender_id"`
	ReceiverID     string `json:"receiver_id,omitempty"` // Deprecated, use ConversationID
	Content        string `json:"content"`
	Timestamp      int64  `json:"timestamp,omitempty"`
}

type MessageDB struct {
	Id             string `bson:"_id,omitempty" json:"id"`
	ConversationID string `bson:"conversation_id" json:"conversation_id"`
	SenderID       string `bson:"sender_id" json:"sender_id"`
	ReceiverID     string `bson:"receiver_id,omitempty" json:"receiver_id,omitempty"` // Deprecated
	Content        string `bson:"content" json:"content"`
	Timestamp      int64  `bson:"timestamp" json:"timestamp"`
}

type MessageResponse struct {
	Id             string   `json:"id"`
	ConversationID string   `json:"conversation_id"`
	SenderID       string   `json:"sender_id"`
	ReceiverID     string   `json:"receiver_id,omitempty"`
	Recipients     []string `json:"recipients,omitempty"` // List of user IDs to receive the message
	Content        string   `json:"content"`
	Timestamp      int64    `json:"timestamp,omitempty"`
}

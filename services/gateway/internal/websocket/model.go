package websocket

import "encoding/json"

type MessageType string

const (
	MessageTypeChat        MessageType = "chat_message"
	MessageTypeChangeChat  MessageType = "change_chat"
	MessageTypeTypingStart MessageType = "typing_start"
	MessageTypeTypingStop  MessageType = "typing_stop"
)

type GatewayMessage struct {
	Type      MessageType     `json:"type"`
	SenderID  string          `json:"sender_id"`
	Timestamp int64           `json:"timestamp,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}

type ChatMessagePayload struct {
	ConversationID string `json:"conversation_id,omitempty"` // New: use conversation
	SenderID       string `json:"sender_id"`
	ReceiverID     string `json:"receiver_id,omitempty"` // Deprecated: for auto-creating conversation
	Content        string `json:"content"`
	Timestamp      int64  `json:"timestamp,omitempty"`
}

type TypingPayload struct {
	ChatID   string `json:"chat_id"`
	UserID   string `json:"user_id"`
	IsTyping bool   `json:"is_typing"`
}

type ChangeChatPayload struct {
	UserID string `json:"user_id"`
	ChatID string `json:"chat_id"`
}

type HeartbeatPayload struct {
	UserID string `json:"user_id"`
	Time   int64  `json:"time"`
}

type MessageResponse struct {
	ID             string   `json:"id"`
	ConversationID string   `json:"conversation_id,omitempty"`
	SenderID       string   `json:"sender_id"`
	ReceiverID     string   `json:"receiver_id,omitempty"` // Deprecated
	Recipients     []string `json:"recipients,omitempty"`  // List of user IDs to receive the message
	Content        string   `json:"content"`
	Timestamp      int64    `json:"timestamp,omitempty"`
}

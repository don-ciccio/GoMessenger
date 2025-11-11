package websocket

type MessageRequest struct {
	SenderID   int    `json:"sender_id"`
	ReceiverID int    `json:"receiver_id"`
	Content    string `json:"content"`
	Timestamp  int64  `json:"timestamp,omitempty"`
}

type MessageResponse struct {
	ID         int    `json:"id"`
	SenderID   int    `json:"sender_id"`
	ReceiverID int    `json:"receiver_id"`
	Content    string `json:"content"`
	Timestamp  int64  `json:"timestamp,omitempty"`
}

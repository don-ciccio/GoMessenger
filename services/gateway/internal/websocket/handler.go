package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"sync"

	"github.com/Miguel-Pezzini/GoMessenger/services/gateway/internal/auth"
	"github.com/gorilla/websocket"
)

type WsHandler struct {
	service  *Service
	clients  map[string]*websocket.Conn
	clientsM sync.RWMutex
}

// Maximum inbound WebSocket message size (32 KB). Prevents memory exhaustion from oversized frames.
const maxMessageSize = 32 * 1024

var allowedOrigins = func() []string {
	env := os.Getenv("CORS_ALLOWED_ORIGINS")
	if env == "" {
		return nil // dev: allow all
	}
	var origins []string
	for _, o := range strings.Split(env, ",") {
		if trimmed := strings.TrimSpace(o); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}()

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		if len(allowedOrigins) == 0 {
			return true // dev fallback
		}
		origin := r.Header.Get("Origin")
		// Non-browser clients (iOS app, server-to-server proxy) don't send Origin.
		// Allow them — the JWT token provides authentication.
		if origin == "" {
			return true
		}
		// Browser clients MUST match the whitelist (prevents CSWSH attacks).
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}
		return false
	},
}

func NewWsHandler(service *Service) *WsHandler {
	return &WsHandler{
		service: service,
		clients: make(map[string]*websocket.Conn),
	}
}

func (h *WsHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Erro ao fazer upgrade:", err)
		return
	}

	// Cap inbound message size to prevent memory exhaustion (DoS)
	conn.SetReadLimit(maxMessageSize)
	userID := r.Context().Value(auth.UserIDKey).(string)
	if userID == "" {
		conn.WriteMessage(websocket.TextMessage, []byte("user query param required"))
		conn.Close()
		return
	}

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	h.clientsM.Lock()
	h.clients[userID] = conn
	h.clientsM.Unlock()

	// Send list of online users to the new client
	h.clientsM.Lock()
	onlineUsers := make([]string, 0, len(h.clients))
	for id := range h.clients {
		onlineUsers = append(onlineUsers, id)
	}
	h.clientsM.Unlock()

	conn.WriteJSON(map[string]interface{}{
		"type":     "online_users",
		"user_ids": onlineUsers,
	})

	// Broadcast online status to others
	h.broadcastStatus(userID, true)

	defer func() {
		h.clientsM.Lock()
		delete(h.clients, userID)
		h.clientsM.Unlock()
		conn.Close()

		// Broadcast offline status
		h.broadcastStatus(userID, false)
	}()

	// Start ping loop to keep connection alive
	go h.startPingLoop(userID, conn)

	// Listen for messages from client
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			log.Println("Conexão encerrada:", err)
			break
		}

		var gatewayMessage GatewayMessage
		if err := json.Unmarshal(msgBytes, &gatewayMessage); err != nil {
			log.Println("Erro ao parsear mensagem:", err)
			continue
		}
		switch gatewayMessage.Type {
		case MessageTypeChat:
			{
				var payload ChatMessagePayload
				json.Unmarshal(gatewayMessage.Payload, &payload)
				payload.Timestamp = time.Now().Unix()
				payload.SenderID = userID // Enforce authenticated user ID
				h.service.PersistMessage(payload)
			}
		case MessageTypeDelivered, MessageTypeSeen:
			{
				var payload InteractionPayload
				if err := json.Unmarshal(gatewayMessage.Payload, &payload); err != nil {
					log.Printf("[Receipts] Failed to decode interaction payload: %v\n", err)
					break
				}

				log.Printf("[Receipts] Received %s from User %s targeting User %s (Conv: %s, Msg: %s)\n", 
					gatewayMessage.Type, userID, payload.TargetUserID, payload.ConversationID, payload.MessageID)

				// Publish to chat.events so chat_service persists the status update
				chatEventsChannel := os.Getenv("REDIS_CHANNEL_CHAT_EVENTS")
				if chatEventsChannel == "" {
					chatEventsChannel = "chat.events"
				}
				event := map[string]string{
					"type":           string(gatewayMessage.Type),
					"actor_user_id":  userID,
					"target_user_id": payload.TargetUserID,
				}
				if payload.MessageID != "" {
					event["message_id"] = payload.MessageID
				}
				if payload.ConversationID != "" {
					event["conversation_id"] = payload.ConversationID
				}
				eventJSON, _ := json.Marshal(event)
				if err := h.service.PublishInteraction(chatEventsChannel, string(eventJSON)); err != nil {
					log.Printf("[Receipts] Failed to publish interaction event to redis: %v\n", err)
				} else {
					log.Printf("[Receipts] Successfully published %s to %s\n", gatewayMessage.Type, chatEventsChannel)
				}

				// Map event type to clean viewed_status value
				viewedStatus := "sent"
				if gatewayMessage.Type == MessageTypeDelivered {
					viewedStatus = "delivered"
				} else if gatewayMessage.Type == MessageTypeSeen {
					viewedStatus = "seen"
				}

				// Forward to the message sender so their checkmarks update in real-time
				statusUpdate := map[string]interface{}{
					"type":          string(gatewayMessage.Type),
					"viewed_status": viewedStatus,
				}
				if payload.MessageID != "" {
					statusUpdate["message_id"] = payload.MessageID
				}
				if payload.ConversationID != "" {
					statusUpdate["conversation_id"] = payload.ConversationID
				}
				h.clientsM.RLock()
				if senderConn, ok := h.clients[payload.TargetUserID]; ok {
					err := senderConn.WriteJSON(statusUpdate)
					log.Printf("[Receipts] Forwarded %s to target User %s (err: %v)\n", gatewayMessage.Type, payload.TargetUserID, err)
				} else {
					log.Printf("[Receipts] Target User %s is not connected, skipped forwarding over WS\n", payload.TargetUserID)
				}
				h.clientsM.RUnlock()
			}
		}
	}

	h.clientsM.Lock()
	delete(h.clients, userID)
	h.clientsM.Unlock()
}

func (h *WsHandler) StartPubSubListener() {
	h.service.SubscribeChatChannel(os.Getenv("REDIS_CHANNEL_CHAT"), func(payload string) {
		var msg MessageResponse
		if err := json.Unmarshal([]byte(payload), &msg); err != nil {
			log.Println("Erro ao parsear mensagem Pub/Sub:", err)
			return
		}

		h.clientsM.Lock()
		defer h.clientsM.Unlock()

		// Send to all recipients (includes sender and other participants)
		for _, recipientID := range msg.Recipients {
			if conn, ok := h.clients[recipientID]; ok {
				conn.WriteJSON(msg)
			}
		}

		// Fallback for backward compatibility or direct messages
		if len(msg.Recipients) == 0 {
			if conn, ok := h.clients[msg.ReceiverID]; ok {
				conn.WriteJSON(msg)
			}
			if conn, ok := h.clients[msg.SenderID]; ok {
				conn.WriteJSON(msg)
			}
		}
	})
}

func (h *WsHandler) startPingLoop(userID string, conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := conn.WriteControl(
			websocket.PingMessage,
			[]byte{},
			time.Now().Add(5*time.Second),
		); err != nil {
			log.Println("Ping error, closing connection:", err)
			conn.Close()
			return
		}
	}
}

func (h *WsHandler) broadcastStatus(userID string, online bool) {
	statusMsg := map[string]interface{}{
		"type":      "user_status",
		"user_id":   userID,
		"online":    online,
		"timestamp": time.Now().Unix(),
	}

	h.clientsM.Lock()
	defer h.clientsM.Unlock()

	for _, conn := range h.clients {
		conn.WriteJSON(statusMsg)
	}
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	chat "github.com/Miguel-Pezzini/GoMessenger/services/chat_service/internal"
	mongoutils "github.com/Miguel-Pezzini/GoMessenger/services/chat_service/internal/mongo"
	redisutil "github.com/Miguel-Pezzini/GoMessenger/services/chat_service/internal/redis"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	addr        string
	rdb         *redis.Client
	mongo       *mongo.Database
	redisConfig *redisutil.RedisConfig
}

func NewServer(addr string) *Server {
	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://localhost:27018"
	}

	mongo, err := mongoutils.NewMongoClient(mongoURL, "chatdb")
	if err != nil {
		log.Fatalf("failed to connecting to chat database: %v", err)
	}
	rdb, err := redisutil.NewRedisClient()
	if err != nil {
		log.Fatal("error connecting with redis", err)
	}
	redisConfig := redisutil.LoadRedisConfig()
	return &Server{addr: addr, rdb: rdb, mongo: mongo, redisConfig: redisConfig}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) Start() error {
	ctx := context.Background()
	chat.InitAPNs()

	// Initialize repositories and services
	messageRepo := chat.NewMongoRepository(s.mongo)
	conversationRepo := chat.NewMongoConversationRepository(s.mongo)
	messageService := chat.NewService(messageRepo)
	conversationService := chat.NewConversationService(conversationRepo)

	// Initialize handlers
	conversationHandler := chat.NewConversationHandler(conversationService, messageRepo)

	// Initialize broadcast handler
	broadcastRepo := chat.NewMongoBroadcastRepository(s.mongo)
	broadcastHandler := chat.NewBroadcastHandler(
		broadcastRepo, conversationService, s.rdb, s.redisConfig.StreamChat)

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("POST /conversations", conversationHandler.CreateOrGetConversation)
	mux.HandleFunc("GET /conversations", conversationHandler.ListConversations)
	mux.HandleFunc("GET /conversations/archived", conversationHandler.ListArchivedConversations)
	mux.HandleFunc("GET /conversations/{id}/messages", conversationHandler.GetConversationMessages)
	mux.HandleFunc("POST /conversations/{id}/archive", conversationHandler.ArchiveConversation)
	mux.HandleFunc("POST /conversations/{id}/unarchive", conversationHandler.UnarchiveConversation)
	mux.HandleFunc("POST /broadcasts", broadcastHandler.CreateBroadcast)
	mux.HandleFunc("GET /broadcasts", broadcastHandler.ListBroadcasts)
	mux.HandleFunc("GET /broadcasts/{id}", broadcastHandler.GetBroadcast)
	mux.HandleFunc("POST /badge/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var payload struct {
			Badge int `json:"badge"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.rdb.Set(ctx, "badge:"+id, payload.Badge, 0)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Start HTTP server in background
	go func() {
		log.Println("Chat Service HTTP API running on port 8081")
		if err := http.ListenAndServe(s.addr, corsMiddleware(mux)); err != nil {
			log.Fatal("HTTP server failed:", err)
		}
	}()

	// Start chat.events Pub/Sub subscriber for delivery/read receipts
	chatEventsChannel := os.Getenv("REDIS_CHANNEL_CHAT_EVENTS")
	if chatEventsChannel == "" {
		chatEventsChannel = "chat.events"
	}
	go func() {
		pubsub := s.rdb.Subscribe(ctx, chatEventsChannel)
		ch := pubsub.Channel()
		log.Println("Subscribed to", chatEventsChannel, "for read receipts")
		for msg := range ch {
			var event chat.InteractionEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Printf("[Receipts] failed to decode interaction event: %v\n", err)
				continue
			}

			log.Printf("[Receipts] Processing %s for Conv %s, Msg %s, TargetUser %s\n", 
				event.Type, event.ConversationID, event.MessageID, event.TargetUserID)

			switch event.Type {
			case "message_delivered":
				if event.MessageID != "" {
					if err := messageRepo.UpdateViewedStatus(ctx, event.MessageID, chat.ViewedStatusDelivered); err != nil {
						log.Printf("[Receipts] failed to update viewed_status to delivered: %v\n", err)
					} else {
						log.Printf("[Receipts] Successfully updated msg %s to delivered\n", event.MessageID)
					}
				}
			case "message_seen":
				// Batch: mark all messages from the sender in this conversation as seen
				if event.ConversationID != "" && event.TargetUserID != "" {
					if err := messageRepo.MarkConversationSeen(ctx, event.ConversationID, event.TargetUserID); err != nil {
						log.Printf("[Receipts] failed to batch-mark conversation as seen: %v\n", err)
					} else {
						log.Printf("[Receipts] Successfully batch-marked conv %s targeting user %s as seen\n", event.ConversationID, event.TargetUserID)
					}
				} else if event.MessageID != "" {
					if err := messageRepo.UpdateViewedStatus(ctx, event.MessageID, chat.ViewedStatusSeen); err != nil {
						log.Printf("[Receipts] failed to update viewed_status to seen: %v\n", err)
					} else {
						log.Printf("[Receipts] Successfully updated msg %s to seen\n", event.MessageID)
					}
				}
			}
		}
	}()

	// Start Redis stream consumer
	streamName := s.redisConfig.StreamChat

	for {
		streams, err := s.rdb.XRead(ctx, &redis.XReadArgs{
			Streams: []string{streamName, "0"},
			Block:   5 * time.Second,
			Count:   10,
		}).Result()
		if err != nil {
			log.Println("XRead failed:", err)
			time.Sleep(time.Second)
			continue
		}

		for _, st := range streams {
			for _, msg := range st.Messages {

				rawData, ok := msg.Values["data"].(string)
				if !ok {
					log.Println("invalid message format, missing 'data'")
					_ = s.rdb.XDel(ctx, streamName, msg.ID).Err()
					continue
				}

				var req chat.MessageRequest
				if err := json.Unmarshal([]byte(rawData), &req); err != nil {
					log.Println("failed to unmarshal message request:", err)
					_ = s.rdb.XDel(ctx, streamName, msg.ID).Err()
					continue
				}

				// Auto-create conversation if not provided
				var conversation *chat.Conversation
				if req.ConversationID == "" && req.ReceiverID != "" {
					participants := []string{req.SenderID, req.ReceiverID}
					var err error
					conversation, err = conversationService.GetOrCreateConversation(ctx, participants, "")
					if err != nil {
						log.Println("failed to create conversation:", err)
						continue
					}
					req.ConversationID = conversation.ID
				} else if req.ConversationID != "" {
					var err error
					conversation, err = conversationService.GetConversation(ctx, req.ConversationID)
					if err != nil {
						log.Println("failed to get conversation:", err)
					}
				}

				messageResponse, err := messageService.Create(ctx, req)
				if err != nil {
					log.Println("failed to persist message:", err)
					continue
				}

				// Update conversation's last message and set recipients
				if req.ConversationID != "" {
					conversationService.UpdateLastMessage(ctx, req.ConversationID, req.Content, req.SenderID)
					if conversation != nil {
						messageResponse.Recipients = conversation.Participants
					}
				}

				res, err := json.Marshal(messageResponse)
				if err != nil {
					log.Println("failed to marshal response:", err)
					continue
				}

				channel := os.Getenv("REDIS_CHANNEL_CHAT")
				if err := s.rdb.Publish(ctx, channel, res).Err(); err != nil {
					log.Println("failed to publish to gateway channel:", err)
					continue
				}

				// Send APNs Push Notification to all recipients except the sender
				if conversation != nil {
					go func(senderID string, text string, conversationID string, participants []string, tag string) {
						var recipientIDs []string
						for _, p := range participants {
							if p != senderID {
								recipientIDs = append(recipientIDs, p)
							}
						}

						if len(recipientIDs) == 0 {
							return
						}

						authServiceHTTPURL := os.Getenv("AUTH_SERVICE_HTTP_URL")
						if authServiceHTTPURL == "" {
							authServiceHTTPURL = "http://localhost:8082"
						}

						// Fetch all participants (including sender) to get display names and device tokens
						allIDs := append(recipientIDs, senderID)
						reqBody, _ := json.Marshal(map[string]interface{}{"ids": allIDs})
						httpReq, _ := http.NewRequest("POST", authServiceHTTPURL+"/users/batch/internal", bytes.NewBuffer(reqBody))
						httpReq.Header.Set("Content-Type", "application/json")

						client := &http.Client{Timeout: 5 * time.Second}
						resp, err := client.Do(httpReq)
						if err != nil {
							log.Println("failed to fetch users for push:", err)
							return
						}
						defer resp.Body.Close()

						if resp.StatusCode == 200 {
							var users []struct {
								ID           string   `json:"id"`
								Username     string   `json:"username"`
								DisplayName  string   `json:"display_name"`
								DeviceTokens []string `json:"device_tokens"`
							}
							if err := json.NewDecoder(resp.Body).Decode(&users); err == nil {
								// Extract sender's name and send pushes individually to recipients
								// Use broadcast tag as push title when present,
								// otherwise fall back to sender's display name.
								senderName := "New Message"
								if tag != "" {
									senderName = tag
								} else {
									for _, u := range users {
										if u.ID == senderID {
											if u.DisplayName != "" {
												senderName = u.DisplayName
											} else {
												senderName = u.Username
											}
										}
									}
								}
								
								for _, u := range users {
									if u.ID != senderID && len(u.DeviceTokens) > 0 {
										// Increment badge for this specific user
										newBadge := s.rdb.Incr(ctx, "badge:"+u.ID).Val()
										
										metadata := map[string]interface{}{
											"conversation_id": conversationID,
										}
										chat.SendPushNotification(u.DeviceTokens, senderName, text, metadata, int(newBadge))
									}
								}
							}
						}
					}(req.SenderID, req.Content, req.ConversationID, conversation.Participants, req.Tag)
				}

				if err := s.rdb.XDel(ctx, streamName, msg.ID).Err(); err != nil {
					log.Println("failed to delete processed stream entry:", err)
				}
			}
		}
	}
}

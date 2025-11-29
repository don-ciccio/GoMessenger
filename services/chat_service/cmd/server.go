package main

import (
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

	// Initialize repositories and services
	messageRepo := chat.NewMongoRepository(s.mongo)
	conversationRepo := chat.NewMongoConversationRepository(s.mongo)
	messageService := chat.NewService(messageRepo)
	conversationService := chat.NewConversationService(conversationRepo)

	// Initialize handlers
	conversationHandler := chat.NewConversationHandler(conversationService, messageRepo)

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("POST /conversations", conversationHandler.CreateOrGetConversation)
	mux.HandleFunc("GET /conversations", conversationHandler.ListConversations)
	mux.HandleFunc("GET /conversations/{id}/messages", conversationHandler.GetConversationMessages)

	// Start HTTP server in background
	go func() {
		log.Println("Chat Service HTTP API running on port 8081")
		if err := http.ListenAndServe(s.addr, corsMiddleware(mux)); err != nil {
			log.Fatal("HTTP server failed:", err)
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
					conversation, err = conversationService.GetOrCreateConversation(ctx, participants)
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
					conversationService.UpdateLastMessage(ctx, req.ConversationID, req.Content)
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

				if err := s.rdb.XDel(ctx, streamName, msg.ID).Err(); err != nil {
					log.Println("failed to delete processed stream entry:", err)
				}
			}
		}
	}
}

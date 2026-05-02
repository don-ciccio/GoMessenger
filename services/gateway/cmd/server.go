package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Miguel-Pezzini/GoMessenger/services/gateway/internal/auth"
	authpb "github.com/Miguel-Pezzini/GoMessenger/services/gateway/internal/pb/auth"
	redisutil "github.com/Miguel-Pezzini/GoMessenger/services/gateway/internal/redis"
	"github.com/Miguel-Pezzini/GoMessenger/services/gateway/internal/websocket"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	addr         string
	rdb          *redis.Client
	auth_service authpb.AuthServiceClient
}

func NewServer(addr, authAddr string) *Server {
	redisClient, err := redisutil.NewRedisClient()
	if err != nil {
		log.Fatal("error connecting with redis", err)
	}

	authServiceAddr := os.Getenv("AUTH_SERVICE_ADDR")
	if authServiceAddr == "" {
		authServiceAddr = authAddr
	}

	authService, err := auth.NewAuthServiceClient(authServiceAddr)
	if err != nil {
		log.Fatal("error connecting with auth service", err)
	}
	log.Println("Gateway connected with Authentication Service")
	return &Server{addr: addr, rdb: redisClient, auth_service: authService}
}

func corsMiddleware(next http.Handler) http.Handler {
	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS") // comma-separated: "https://admin.shopify.com,https://bitemein.app,https://noboringcoffee.com"
	if allowedOrigins == "" {
		allowedOrigins = "*" // dev fallback
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if allowedOrigins == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			// Check if the request origin is in the whitelist
			for _, allowed := range splitOrigins(allowedOrigins) {
				if origin == allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func splitOrigins(s string) []string {
	var origins []string
	for _, o := range strings.Split(s, ",") {
		if trimmed := strings.TrimSpace(o); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

// proxyToConversationService forwards requests to chat_service
func proxyToConversationService(w http.ResponseWriter, r *http.Request) {
	chatServiceURL := os.Getenv("CHAT_SERVICE_URL")
	if chatServiceURL == "" {
		chatServiceURL = "http://localhost:8081"
	}

	// Build target URL
	targetURL := chatServiceURL + r.URL.Path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	// Create proxy request
	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}

	// Copy headers
	proxyReq.Header = r.Header.Clone()

	// Inject authenticated User ID
	if userID, ok := r.Context().Value(auth.UserIDKey).(string); ok {
		proxyReq.Header.Set("X-User-Id", userID)
	}

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Failed to proxy request", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	io.Copy(w, resp.Body)
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	wsHandler := websocket.NewWsHandler(websocket.NewService(websocket.NewRedisRepository(s.rdb)))

	wsHandler.StartPubSubListener()
	mux.Handle("GET /ws", auth.JWTMiddleware(http.HandlerFunc(wsHandler.HandleConnection)))

	authHandler := auth.NewHandler(auth.NewService(s.auth_service))
	mux.Handle("POST /auth/login", http.HandlerFunc(authHandler.LoginHandler))
	mux.Handle("POST /auth/register", http.HandlerFunc(authHandler.RegisterHandler))

	// Conversation endpoints (proxy to chat_service) - Protected by JWT
	mux.Handle("POST /conversations", auth.JWTMiddleware(http.HandlerFunc(proxyToConversationService)))
	mux.Handle("GET /conversations", auth.JWTMiddleware(http.HandlerFunc(proxyToConversationService)))
	mux.Handle("GET /conversations/{id}/messages", auth.JWTMiddleware(http.HandlerFunc(proxyToConversationService)))

	// Proxy user search and batch lookup to Auth Service - Protected by JWT
	authServiceHTTPURL := os.Getenv("AUTH_SERVICE_HTTP_URL")
	if authServiceHTTPURL == "" {
		authServiceHTTPURL = "http://localhost:8082"
	}
	mux.Handle("GET /users/search", auth.JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyToService(w, r, authServiceHTTPURL)
	})))
	mux.Handle("POST /users/batch", auth.JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyToService(w, r, authServiceHTTPURL)
	})))
	mux.Handle("POST /users/device-token", auth.JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyToService(w, r, authServiceHTTPURL)
	})))
	mux.Handle("DELETE /users/device-token", auth.JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyToService(w, r, authServiceHTTPURL)
	})))

	return http.ListenAndServe(s.addr, corsMiddleware(mux))
}

// Helper function for proxying to any service
func proxyToService(w http.ResponseWriter, r *http.Request, baseURL string) {
	targetURL := baseURL + r.URL.Path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}

	proxyReq.Header = r.Header.Clone()

	// Inject authenticated User ID
	if userID, ok := r.Context().Value(auth.UserIDKey).(string); ok {
		proxyReq.Header.Set("X-User-Id", userID)
	}

	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Failed to proxy request", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)
}

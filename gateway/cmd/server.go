package main

import (
	"context"
	"log"
	"net/http"
	"real_time_chat/internal/auth"
	"real_time_chat/internal/websocket"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	addr string
}

func NewServer(addr string) *Server {
	return &Server{addr: addr}
}

func (s *Server) Start() error {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	db := client.Database("chatdb")

	mux := http.NewServeMux()
	mux.Handle("GET /ws", http.HandlerFunc(websocket.WsHandler))

	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authService)

	mux.Handle("POST /auth/login", http.HandlerFunc(authHandler.LoginHandler))
	mux.Handle("POST /auth/register", http.HandlerFunc(authHandler.RegisterHandler))
	return http.ListenAndServe(s.addr, mux)
}

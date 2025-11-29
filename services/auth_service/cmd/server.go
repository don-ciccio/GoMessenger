package main

import (
	"context"
	"os"

	auth "github.com/Miguel-Pezzini/GoMessenger/services/auth_service/internal"
	authpb "github.com/Miguel-Pezzini/GoMessenger/services/auth_service/internal/pb"
)

type Server struct {
	authpb.UnimplementedAuthServiceServer
	service *auth.Service
	repo    auth.Repository
}

func NewServer() (*Server, error) {
	mongoURL := os.Getenv("MONGO_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://localhost:27019"
	}

	mongoDB, err := NewMongoClient(mongoURL, "userdb")
	if err != nil {
		return nil, err
	}
	repo := auth.NewMongoRepository(mongoDB)
	svc := auth.NewService(repo)
	return &Server{service: svc, repo: repo}, nil
}

func (s *Server) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	return s.service.Register(req)
}

func (s *Server) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	return s.service.Authenticate(req)
}

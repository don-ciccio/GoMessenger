package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	auth "github.com/Miguel-Pezzini/GoMessenger/services/auth_service/internal"
	authpb "github.com/Miguel-Pezzini/GoMessenger/services/auth_service/internal/pb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

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

func main() {
	srv, err := NewServer()
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	// Start HTTP server for search endpoint in background
	go func() {
		mux := http.NewServeMux()
		searchHandler := auth.NewSearchHandler(srv.repo)
		mux.HandleFunc("GET /users/search", searchHandler.SearchUsers)
		mux.HandleFunc("POST /users/batch", searchHandler.GetUsers)

		log.Println("AuthService HTTP API running on port 8082")
		if err := http.ListenAndServe(":8082", corsMiddleware(mux)); err != nil {
			log.Fatal("HTTP server failed:", err)
		}
	}()

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	authpb.RegisterAuthServiceServer(grpcServer, srv)

	log.Println("AuthService rodando na porta 50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}

func NewMongoClient(URI, dbName string) (*mongo.Database, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(URI))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	return client.Database(dbName), nil
}

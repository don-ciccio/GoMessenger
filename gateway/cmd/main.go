package main

import (
	"log"

	"github.com/Miguel-Pezzini/real_time_chat/gateway/internal/auth"
	db "github.com/Miguel-Pezzini/real_time_chat/pkg/db"
)

func main() {
	redisClient, err := db.NewRedisClient()
	if err != nil {
		log.Fatal("error connecting with redis", err)
	}
	authService, err := auth.NewAuthServiceClient("localhost:50051")
	if err != nil {
		log.Fatal("error connecting with auth service", err)
	}
	log.Println("Gateway connected with Authentication Service")

	server := NewServer(":8080", redisClient, authService)
	log.Println("Gateway running on port 8080")
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}

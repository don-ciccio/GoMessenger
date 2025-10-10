package main

import (
	"log"

	db "github.com/Miguel-Pezzini/real_time_chat/pkg/db"
)

func main() {
	mongoDB, err := db.NewMongoClient("mongodb://localhost:27017", "chatdb")
	if err != nil {
		log.Fatal(err)
	}
	redisClient := db.NewRedisClient()

	server := NewServer(":8080", mongoDB, redisClient)
	log.Println("Gateway running on port 8080")
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"log"
	"real_time_chat/pkg/db"
)

func main() {
	a, err := db.NewMongoClient("mongodb://localhost:27017", "userdb")
	if err != nil {
		log.Fatal(err)
	}

	server := NewServer(":8080", db)
	log.Println("Gateway running on port 8080")
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}

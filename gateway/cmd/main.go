package main

import (
	"log"
)

func main() {
	server := NewServer(":8080")
	log.Println("Gateway running on port 8080")
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}

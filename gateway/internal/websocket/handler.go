package websocket

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type WsHandler struct {
	service *Service
}

var (
	ctx      = context.Background()
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	clients   = make(map[string]*websocket.Conn)
	clientsMu sync.RWMutex
)

func NewWsHandler(service *Service) *WsHandler {
	return &WsHandler{service: service}
}

func (ws *WsHandler) WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

package main

import (
	"github.com/gorilla/websocket"
	"log"
)

func main() {
	dial, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Println("Failed to connect to WebSocket server:", err)
	}
	dial.WriteMessage(websocket.TextMessage, []byte("Hello, WebSocket!"))
	for {
		_, p, err := dial.ReadMessage()
		if err != nil {

			log.Println("Failed to read message:", err)
		}
		log.Println("Received: ", string(p))
	}
}

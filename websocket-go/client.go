package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type client struct {
	socket *websocket.Conn
	send   chan []byte
	room   *room
}

func (c *client) read() {
	log.Println("info: [client] Begin to observe WebSocket.")
	for {
		if _, msg, err := c.socket.ReadMessage(); err == nil {
			log.Println("info: [client] Received a message.")
			c.room.forward <- msg
		} else {
			break
		}
	}
	c.socket.Close()
	log.Println("info: [client] Closed WebSocket.")
}

func (c *client) write() {
	log.Println("info: [client] Begin to monitor write queue.")
	for msg := range c.send {
		log.Println("info: [client] Sending a message.")
		if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
	c.socket.Close()
	log.Println("info: [client] Close WebSocket.")
}

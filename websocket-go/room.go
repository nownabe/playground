package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

type room struct {
	forward chan []byte
	join    chan *client
	leave   chan *client
	clients map[*client]bool
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Println("info: Connecting WebSocket...")
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	log.Println("info: Joinning room...")
	r.join <- client
	defer func() {
		log.Println("info: Lefting room...")
		r.leave <- client
	}()
	go client.write()
	client.read()
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			log.Println("info: [room] A client joined room.")
			r.clients[client] = true
		case client := <-r.leave:
			log.Println("info: [room] A client left room.")
			delete(r.clients, client)
			close(client.send)
		case msg := <-r.forward:
			log.Println("info: [room] Received message.")
			for client := range r.clients {
				select {
				case client.send <- msg:
					log.Println("info: [room] Sent message.")
				default:
					log.Println("error: [room] Failed to send message.")
					delete(r.clients, client)
					close(client.send)
				}
			}
		}
	}
}

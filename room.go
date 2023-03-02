package main

import (
	"chicko_chat/log"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize,
	WriteBufferSize: socketBufferSize}

type room struct {
	// channel that holds incoming messages
	forward chan []byte

	join chan *client

	leave chan *client

	clients map[*client]bool

	logger logger.Logger
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
			r.logger.Log("New Client Joined")

		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
			r.logger.Log("Client Left")

		case msg := <-r.forward:
			r.logger.Log("Message received: ", string(msg))
			for client := range r.clients {
				client.send <- msg
				r.logger.Log("Sent to clients")
			}
		}
	}
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}

func NewRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}
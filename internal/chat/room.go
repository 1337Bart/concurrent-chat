package chat

import (
	"log"
	"sync"
)

type Room struct {
	ID        string
	clients   map[*Client]bool
	broadcast chan []byte
	join      chan *Client
	leave     chan *Client
	mu        sync.Mutex
}

func NewRoom(id string) *Room {
	return &Room{
		ID:        id,
		clients:   make(map[*Client]bool),
		broadcast: make(chan []byte),
		join:      make(chan *Client),
		leave:     make(chan *Client),
	}
}

func (r *Room) Run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
		case client := <-r.leave:
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
			}
		case message := <-r.broadcast:
			for client := range r.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(r.clients, client)
				}
			}
		}
	}
}

func (r *Room) Join(client *Client) {
	log.Printf("Client joining room %s", r.ID)
	r.join <- client
}

func (r *Room) Leave(client *Client) {
	log.Printf("Client leaving room %s", r.ID)
	r.leave <- client
}

func (r *Room) Broadcast(message []byte) {
	log.Printf("Received message for broadcast in room %s: %s", r.ID, string(message))
	r.broadcast <- message
}

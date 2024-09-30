package chat

import (
	"log"
	"sync"
)

type Room struct {
	ID        string
	clients   map[*Client]bool
	join      chan *Client
	leave     chan *Client
	broadcast chan []byte
	mu        sync.Mutex
}

func NewRoom(id string) *Room {
	return &Room{
		ID:        id,
		clients:   make(map[*Client]bool),
		join:      make(chan *Client),
		leave:     make(chan *Client),
		broadcast: make(chan []byte),
	}
}

func (r *Room) Run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
			log.Printf("Client joined room: %s", r.ID)
		case client := <-r.leave:
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.send)
				log.Printf("Client left room: %s", r.ID)
			}
		case message := <-r.broadcast:
			log.Printf("Broadcasting message in room %s: %s", r.ID, string(message))
			for client := range r.clients {
				select {
				case client.send <- message:
					log.Printf("Sent message to client in room %s", r.ID)
				default:
					close(client.send)
					delete(r.clients, client)
					log.Printf("Removed unresponsive client from room %s", r.ID)
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

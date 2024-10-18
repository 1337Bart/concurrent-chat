package chat

import (
	"log"
	"sync"
)

type Room struct {
	ID           string
	participants map[*ChatParticipant]bool
	broadcast    chan []byte
	join         chan *ChatParticipant
	leave        chan *ChatParticipant
	mu           sync.Mutex
}

func NewRoom(id string) *Room {
	return &Room{
		ID:           id,
		participants: make(map[*ChatParticipant]bool),
		broadcast:    make(chan []byte),
		join:         make(chan *ChatParticipant),
		leave:        make(chan *ChatParticipant),
	}
}

func (r *Room) Run() {
	for {
		select {
		case participant := <-r.join:
			r.participants[participant] = true
		case participant := <-r.leave:
			if _, ok := r.participants[participant]; ok {
				delete(r.participants, participant)
			}
		case message := <-r.broadcast:
			for participant := range r.participants {
				select {
				case participant.Conn.send <- message:
				default:
					close(participant.Conn.send)
					delete(r.participants, participant)
				}
			}
		}
	}
}

func (r *Room) Join(participant *ChatParticipant) {
	r.join <- participant
}

func (r *Room) Leave(participant *ChatParticipant) {
	r.leave <- participant
	log.Printf("Participant queued to leave room %s", r.ID)
}

func (r *Room) Broadcast(message []byte) {
	r.broadcast <- message
}

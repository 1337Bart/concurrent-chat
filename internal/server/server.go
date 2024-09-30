package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"chat/internal/chat"
	"chat/internal/config"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Server struct {
	config *config.Config
	router *mux.Router
	rooms  map[string]*chat.Room
	mu     sync.RWMutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now
	},
}

func NewServer(cfg *config.Config) *Server {
	s := &Server{
		config: cfg,
		router: mux.NewRouter(),
		rooms:  make(map[string]*chat.Room),
	}
	s.routes()
	return s
}

func (s *Server) Router() *mux.Router {
	return s.router
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received WebSocket connection request from %s", r.RemoteAddr)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}

	log.Printf("WebSocket connection established")

	// Handle the connection (you'll need to implement this)
	go s.handleConnection(conn)
}

func (s *Server) handleConnection(conn *websocket.Conn) {
	var currentRoom string
	var currentClient *chat.Client

	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			if currentRoom != "" && currentClient != nil {
				s.mu.RLock()
				room, exists := s.rooms[currentRoom]
				s.mu.RUnlock()
				if exists {
					room.Leave(currentClient)
				}
			}
			return
		}
		var message map[string]interface{}
		if err := json.Unmarshal(p, &message); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		messageType, ok := message["type"].(string)
		if !ok {
			log.Printf("Message type not found or not a string")
			continue
		}

		switch messageType {
		case "chat":
			roomName, ok := message["room"].(string)
			if !ok {
				log.Printf("Room not found in chat message")
				continue
			}
			s.mu.RLock()
			room, exists := s.rooms[roomName]
			s.mu.RUnlock()
			if !exists {
				log.Printf("Room %s does not exist", roomName)
				continue
			}
			log.Printf("Received message for broadcast in room %s: %s", roomName, string(p))
			room.Broadcast(p)
		case "join":
			roomName, ok := message["room"].(string)
			if !ok {
				log.Printf("Room not found in join message")
				continue
			}
			s.mu.Lock()
			room, exists := s.rooms[roomName]
			if !exists {
				room = chat.NewRoom(roomName)
				s.rooms[roomName] = room
				go room.Run()
			}
			s.mu.Unlock()
			if currentRoom != "" && currentClient != nil {
				s.mu.RLock()
				oldRoom, exists := s.rooms[currentRoom]
				s.mu.RUnlock()
				if exists {
					oldRoom.Leave(currentClient)
				}
			}
			currentRoom = roomName
			currentClient = chat.NewClient(conn, room)
			room.Join(currentClient)
			go currentClient.WritePump()
			log.Printf("Client joined room: %s", roomName)
		case "leave":
			if currentRoom != "" && currentClient != nil {
				s.mu.RLock()
				room, exists := s.rooms[currentRoom]
				s.mu.RUnlock()
				if exists {
					room.Leave(currentClient)
					log.Printf("Client left room: %s", currentRoom)
				}
				currentRoom = ""
				currentClient = nil
			}
		default:
			log.Printf("Unknown message type: %s", messageType)
		}
	}
}

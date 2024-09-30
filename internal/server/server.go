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

type ClientConnection struct {
	conn   *websocket.Conn
	rooms  map[string]*chat.Client
	server *Server
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

	go s.handleConnection(conn)
}
func (s *Server) handleConnection(conn *websocket.Conn) {
	client := chat.NewClient(conn)

	client.HandleMessage = func(message []byte) {
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			return
		}

		messageType, ok := msg["type"].(string)
		if !ok {
			log.Printf("Message type not found or not a string")
			return
		}

		switch messageType {
		case "chat":
			roomName, ok := msg["room"].(string)
			if !ok {
				log.Printf("Room not found in chat message")
				return
			}
			room, exists := client.Rooms[roomName]
			if !exists {
				log.Printf("Client not in room %s", roomName)
				return
			}
			log.Printf("Received message for broadcast in room %s: %s", roomName, string(message))
			room.Broadcast(message)
		case "join":
			roomName, ok := msg["room"].(string)
			if !ok {
				log.Printf("Room not found in join message")
				return
			}
			s.mu.Lock()
			room, exists := s.rooms[roomName]
			if !exists {
				room = chat.NewRoom(roomName)
				s.rooms[roomName] = room
				go room.Run()
			}
			s.mu.Unlock()
			client.Rooms[roomName] = room
			room.Join(client)
			log.Printf("Client joined room: %s", roomName)
		case "leave":
			roomName, ok := msg["room"].(string)
			if !ok {
				log.Printf("Room not found in leave message")
				return
			}
			room, exists := client.Rooms[roomName]
			if exists {
				room.Leave(client)
				delete(client.Rooms, roomName)
				log.Printf("Client left room: %s", roomName)
			}
		default:
			log.Printf("Unknown message type: %s", messageType)
		}
	}

	go client.WritePump()
	client.ReadPump() // this will block until the connection is closed
}

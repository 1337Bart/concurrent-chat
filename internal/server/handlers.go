package server

import (
	"encoding/json"
	"log"
	"net/http"

	"chat/internal/chat"
	"github.com/gorilla/mux"
)

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}

	connection := chat.NewConnection(conn)
	participant := chat.NewChatParticipant(connection)

	go s.handleParticipant(participant)
}

func (s *Server) handleParticipant(participant *chat.ChatParticipant) {
	participant.Conn.HandleMessage = func(message []byte) {
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
			room, exists := participant.Rooms[roomName]
			if !exists {
				log.Printf("Participant not in room %s", roomName)
				return
			}
			log.Printf("received message for broadcast: %s", string(message))
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
			participant.JoinRoom(room)
			log.Printf("Participant joined room: %s", roomName)
		case "leave":
			roomName, ok := msg["room"].(string)
			if !ok {
				log.Printf("Room not found in leave message")
				return
			}
			participant.LeaveRoom(roomName)
			log.Printf("Participant left room: %s", roomName)
		default:
			log.Printf("Unknown message type: %s", messageType)
		}
	}

	go participant.Conn.WritePump()
	onClose := func() {
		for roomID, _ := range participant.Rooms {
			participant.LeaveRoom(roomID)
			log.Printf("Participant left room %s due to connection close", roomID)
		}
	}

	participant.Conn.ReadPump(onClose)
}

func (s *Server) handleRoomCreation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomID"]

	log.Printf("Received request to create room: %s", roomID)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.rooms[roomID]; exists {
		log.Printf("Room %s already exists", roomID)
		http.Error(w, "Room already exists", http.StatusConflict)
		return
	}

	room := chat.NewRoom(roomID)
	s.rooms[roomID] = room
	go room.Run()

	log.Printf("Room %s created successfully", roomID)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Room created successfully"))
}

func (s *Server) handleListRooms(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rooms := make([]string, 0, len(s.rooms))
	for roomID := range s.rooms {
		rooms = append(rooms, roomID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

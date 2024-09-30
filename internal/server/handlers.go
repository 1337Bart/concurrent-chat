package server

import (
	"encoding/json"
	"log"
	"net/http"

	"chat/internal/chat"
	"github.com/gorilla/mux"
)

func (s *Server) routes() {
	s.router.HandleFunc("/ws", s.handleWebSocket) // Change this line
	s.router.HandleFunc("/room/{roomID}", s.handleRoomCreation).Methods("POST")
	s.router.HandleFunc("/rooms", s.handleListRooms).Methods("GET")
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

package server

import "github.com/gorilla/mux"

func (s *Server) Router() *mux.Router {
	return s.router
}

func (s *Server) routes() {
	s.router.HandleFunc("/ws", s.handleWebSocket)
	s.router.HandleFunc("/room/{roomID}", s.handleRoomCreation).Methods("POST")
	s.router.HandleFunc("/rooms", s.handleListRooms).Methods("GET")
}

package server

import (
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
	rooms  map[string]*chat.Connection
	server *Server
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
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

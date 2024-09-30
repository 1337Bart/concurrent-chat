package main

import (
	"log"
	"net/http"

	"chat/internal/config"
	"chat/internal/server"
)

func main() {
	log.Println("Starting chat application")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	s := server.NewServer(cfg)

	// Serve static files
	fs := http.FileServer(http.Dir("./web/static"))
	s.Router().PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	log.Println("Static file server set up")

	// Serve index.html
	s.Router().HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Serving index.html")
		http.ServeFile(w, r, "./web/templates/index.html")
	})

	log.Printf("Starting server on %s", cfg.Address)
	if err := http.ListenAndServe(cfg.Address, s.Router()); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

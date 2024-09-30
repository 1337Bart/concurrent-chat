package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"chat/internal/config"
	"chat/internal/server"
)

func TestListRooms(t *testing.T) {
	cfg := &config.Config{Address: ":8080"}
	s := server.NewServer(cfg)

	// Create a test server
	ts := httptest.NewServer(s.Router())
	defer ts.Close()

	// Create a room first
	_, err := http.Post(ts.URL+"/room/testroom", "", nil)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// Test listing rooms
	resp, err := http.Get(ts.URL + "/rooms")
	if err != nil {
		t.Fatalf("Failed to get rooms: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}

	var rooms []string
	if err := json.NewDecoder(resp.Body).Decode(&rooms); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(rooms) != 1 || rooms[0] != "testroom" {
		t.Errorf("Unexpected rooms list: %v", rooms)
	}
}

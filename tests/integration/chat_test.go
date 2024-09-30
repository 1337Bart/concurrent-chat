package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"chat/internal/config"
	"chat/internal/server"
)

func TestCreateRoom(t *testing.T) {
	cfg := &config.Config{Address: ":8080"}
	s := server.NewServer(cfg)

	ts := httptest.NewServer(s.Router())
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/room/testroom", "", nil)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status Created, got %v", resp.Status)
	}
}

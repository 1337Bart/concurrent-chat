package unit

import (
	"testing"
	"time"

	"chat/internal/chat"
)

func TestMessageEncodeDecode(t *testing.T) {
	original := &chat.Message{
		Type:      "text",
		Content:   "Hello, World!",
		Sender:    "Alice",
		Timestamp: time.Now(),
	}

	encoded, err := original.Encode()
	if err != nil {
		t.Fatalf("Failed to encode message: %v", err)
	}

	decoded, err := chat.DecodeMessage(encoded)
	if err != nil {
		t.Fatalf("Failed to decode message: %v", err)
	}

	if decoded.Type != original.Type {
		t.Errorf("Type mismatch. Got %s, want %s", decoded.Type, original.Type)
	}
	if decoded.Content != original.Content {
		t.Errorf("Content mismatch. Got %s, want %s", decoded.Content, original.Content)
	}
	if decoded.Sender != original.Sender {
		t.Errorf("Sender mismatch. Got %s, want %s", decoded.Sender, original.Sender)
	}
	if !decoded.Timestamp.Equal(original.Timestamp) {
		t.Errorf("Timestamp mismatch. Got %v, want %v", decoded.Timestamp, original.Timestamp)
	}
}

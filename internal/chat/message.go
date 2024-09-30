package chat

import (
	"encoding/json"
	"time"
)

type Message struct {
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Sender    string    `json:"sender"`
	Timestamp time.Time `json:"timestamp"`
}

func NewMessage(messageType, content, sender string) *Message {
	return &Message{
		Type:      messageType,
		Content:   content,
		Sender:    sender,
		Timestamp: time.Now(),
	}
}

func (m *Message) Encode() ([]byte, error) {
	return json.Marshal(m)
}

func DecodeMessage(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

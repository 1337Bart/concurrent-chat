package chat

import (
	"log"

	"github.com/gorilla/websocket"
)

const (
	writeBufferSize = 256
)

type Client struct {
	conn *websocket.Conn
	room *Room
	send chan []byte
}

func NewClient(conn *websocket.Conn, room *Room) *Client {
	return &Client{
		conn: conn,
		room: room,
		send: make(chan []byte, writeBufferSize),
	}
}
func (c *Client) ReadPump() {
	defer func() {
		c.room.Leave(c)
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		log.Printf("Received message from client in room %s: %s", c.room.ID, string(message))
		c.room.Broadcast(message)
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
			log.Printf("Sent message to client in room %s: %s", c.room.ID, string(message))
		}
	}
}

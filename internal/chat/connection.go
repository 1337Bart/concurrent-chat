package chat

import (
	"log"

	"github.com/gorilla/websocket"
)

type Connection struct {
	conn          *websocket.Conn
	Rooms         map[string]*Room
	send          chan []byte
	HandleMessage func(message []byte)
}

func NewConnection(conn *websocket.Conn) *Connection {
	return &Connection{
		conn:          conn,
		Rooms:         make(map[string]*Room),
		send:          make(chan []byte, 256),
		HandleMessage: func(message []byte) {},
	}
}

func (c *Connection) ReadPump(onClose func()) {
	defer func() {
		onClose()
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
		c.HandleMessage(message)
	}
}

func (c *Connection) WritePump() {
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
		}
	}
}

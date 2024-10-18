package chat

import "log"

type ChatParticipant struct {
	Conn  *Connection
	Rooms map[string]*Room
}

func NewChatParticipant(conn *Connection) *ChatParticipant {
	return &ChatParticipant{
		Conn:  conn,
		Rooms: make(map[string]*Room),
	}
}

func (cp *ChatParticipant) JoinRoom(room *Room) {
	cp.Rooms[room.ID] = room
	room.Join(cp)
}

func (cp *ChatParticipant) LeaveRoom(roomID string) {
	if room, ok := cp.Rooms[roomID]; ok {
		room.Leave(cp)
		delete(cp.Rooms, roomID)
		log.Printf("Participant left room %s", roomID)
	}
}

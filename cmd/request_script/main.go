package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

var (
	addr            = flag.String("addr", "localhost:8080", "http service address")
	numChats        = flag.Int("chats", 5, "number of concurrent chats")
	numConnsPerChat = flag.Int("conns", 3, "number of connections per chat")
	duration        = flag.Duration("duration", 30*time.Second, "duration of the simulation")
	totalSent       int64
	totalReceived   int64
	startTime       time.Time
)

type Message struct {
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Sender    string    `json:"sender"`
	Timestamp time.Time `json:"timestamp"`
	Room      string    `json:"room"`
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	startTime = time.Now()

	var wg sync.WaitGroup
	done := make(chan struct{})

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for i := 0; i < *numChats; i++ {
		wg.Add(1)
		go runChat(i, &wg, done)
	}

	select {
	case <-time.After(*duration):
		log.Println("Simulation duration reached")
	case sig := <-signalChan:
		log.Printf("Received signal: %v", sig)
	}

	close(done)
	wg.Wait()

	time.Sleep(5 * time.Second) // wait for any lagging msgs
	printFinalStats(time.Since(startTime))
}

func runChat(id int, wg *sync.WaitGroup, done chan struct{}) {
	defer wg.Done()

	var connWg sync.WaitGroup
	roomName := fmt.Sprintf("room_%d", id)

	for j := 0; j < *numConnsPerChat; j++ {
		connWg.Add(1)
		go func(connID int) {
			defer connWg.Done()
			runConnection(id, connID, roomName, done)
		}(j)
	}

	connWg.Wait()
}

func runConnection(chatID, connID int, roomName string, done chan struct{}) {
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Printf("Chat %d, Conn %d connecting to %s", chatID, connID, u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("Chat %d, Conn %d dial error: %v", chatID, connID, err)
		return
	}
	defer c.Close()

	joinRoom(c, roomName)

	go receiveMessages(chatID, connID, c, done)

	ticker := time.NewTicker(time.Duration(rand.Intn(5000)+1000) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			sendMessage(chatID, connID, c, roomName)
			ticker.Reset(time.Duration(rand.Intn(5000)+1000) * time.Millisecond)
		}
	}
}

func receiveMessages(chatID, connID int, c *websocket.Conn, done chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("Chat %d, Conn %d read error: %v", chatID, connID, err)
				return
			}
			atomic.AddInt64(&totalReceived, 1)
			log.Printf("Chat %d, Conn %d received: %s", chatID, connID, message)
		}
	}
}

func sendMessage(chatID, connID int, c *websocket.Conn, roomName string) {
	message := Message{
		Type:      "chat",
		Content:   fmt.Sprintf("Message from chat %d, conn %d at %v", chatID, connID, time.Now()),
		Sender:    fmt.Sprintf("Client_%d_%d", chatID, connID),
		Timestamp: time.Now(),
		Room:      roomName,
	}
	err := c.WriteJSON(message)
	if err != nil {
		log.Printf("Chat %d, Conn %d write error: %v", chatID, connID, err)
		return
	}
	atomic.AddInt64(&totalSent, 1)
}

func joinRoom(c *websocket.Conn, roomName string) {
	joinMessage := Message{
		Type: "join",
		Room: roomName,
	}
	err := c.WriteJSON(joinMessage)
	if err != nil {
		log.Println("join room:", err)
	}
}

func printFinalStats(duration time.Duration) {
	sent := atomic.LoadInt64(&totalSent)
	received := atomic.LoadInt64(&totalReceived)

	fmt.Printf("\n--- Final Statistics ---\n")
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Total Chats: %d\n", *numChats)
	fmt.Printf("Connections per Chat: %d\n", *numConnsPerChat)
	fmt.Printf("Total Connections: %d\n", *numChats**numConnsPerChat)
	fmt.Printf("Messages Sent: %d\n", sent)
	fmt.Printf("Messages Received: %d\n", received)
	fmt.Printf("Send Rate: %.2f messages/second\n", float64(sent)/duration.Seconds())
	fmt.Printf("Receive Rate: %.2f messages/second\n", float64(received)/duration.Seconds())
	fmt.Printf("Average Round Trip Time: %.2f ms\n", float64(duration.Milliseconds())/float64(sent+received)*2)
}

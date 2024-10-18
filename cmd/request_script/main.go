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
	addr          = flag.String("addr", "localhost:8080", "http service address")
	numChats      = flag.Int("chats", 5, "number of concurrent chats")
	duration      = flag.Duration("duration", 1*time.Minute, "duration of the simulation")
	totalSent     int64
	totalReceived int64
	startTime     time.Time
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

	printFinalStats(time.Since(startTime))
}

func runChat(id int, wg *sync.WaitGroup, done chan struct{}) {
	defer wg.Done()

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Printf("Chat %d connecting to %s", id, u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("Chat %d dial error: %v", id, err)
		return
	}
	defer c.Close()

	roomName := fmt.Sprintf("room_%d", id)
	joinRoom(c, roomName)

	go receiveMessages(id, c, done)

	ticker := time.NewTicker(time.Duration(rand.Intn(5000)+1000) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			sendMessage(id, c, roomName)
			ticker.Reset(time.Duration(rand.Intn(5000)+1000) * time.Millisecond)
		}
	}
}

func receiveMessages(id int, c *websocket.Conn, done chan struct{}) {
	for {
		select {
		case <-done:
			return
		default:
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Printf("Chat %d read error: %v", id, err)
				return
			}
			atomic.AddInt64(&totalReceived, 1)
			log.Printf("Chat %d received: %s", id, message)
		}
	}
}

func sendMessage(id int, c *websocket.Conn, roomName string) {
	message := Message{
		Type:      "chat",
		Content:   fmt.Sprintf("Message from chat %d at %v", id, time.Now()),
		Sender:    fmt.Sprintf("Client_%d", id),
		Timestamp: time.Now(),
		Room:      roomName,
	}
	err := c.WriteJSON(message)
	if err != nil {
		log.Printf("Chat %d write error: %v", id, err)
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
	fmt.Printf("Messages Sent: %d\n", sent)
	fmt.Printf("Messages Received: %d\n", received)
	fmt.Printf("Send Rate: %.2f messages/second\n", float64(sent)/duration.Seconds())
	fmt.Printf("Receive Rate: %.2f messages/second\n", float64(received)/duration.Seconds())
	fmt.Printf("Average Round Trip Time: %.2f ms\n", float64(duration.Milliseconds())/float64(sent+received)*2)
}

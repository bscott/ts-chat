package chat

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Message represents a chat message
type Message struct {
	From      string
	Content   string
	Timestamp time.Time
	IsSystem  bool
	IsAction  bool
}

// Room represents a chat room
type Room struct {
	Name      string
	MaxUsers  int
	clients   map[string]*Client
	broadcast chan Message
	join      chan *Client
	leave     chan *Client
	mu        sync.RWMutex
}

// NewRoom creates a new chat room
func NewRoom(name string, maxUsers int) *Room {
	room := &Room{
		Name:      name,
		MaxUsers:  maxUsers,
		clients:   make(map[string]*Client),
		broadcast: make(chan Message),
		join:      make(chan *Client),
		leave:     make(chan *Client),
	}
	
	go room.run()
	return room
}

// run handles room events
func (r *Room) run() {
	for {
		select {
		case client := <-r.join:
			r.addClient(client)
		case client := <-r.leave:
			r.removeClient(client)
		case msg := <-r.broadcast:
			r.broadcastMessage(msg)
		}
	}
}

// addClient adds a client to the room
func (r *Room) addClient(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Check if room is full
	if len(r.clients) >= r.MaxUsers {
		c.sendSystemMessage("Sorry, the room is full. Try again later.")
		c.conn.Close()
		return
	}
	
	// Add client to the room
	r.clients[c.Nickname] = c
	
	// Notify everyone that a new user has joined
	systemMsg := Message{
		From:      "System",
		Content:   fmt.Sprintf("%s has joined the room", c.Nickname),
		Timestamp: time.Now(),
		IsSystem:  true,
	}
	r.broadcastMessage(systemMsg)
}

// removeClient removes a client from the room
func (r *Room) removeClient(c *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.clients[c.Nickname]; exists {
		delete(r.clients, c.Nickname)
		
		// Notify everyone that a user has left
		systemMsg := Message{
			From:      "System",
			Content:   fmt.Sprintf("%s has left the room", c.Nickname),
			Timestamp: time.Now(),
			IsSystem:  true,
		}
		r.broadcastMessage(systemMsg)
	}
}

// broadcastMessage sends a message to all clients
func (r *Room) broadcastMessage(msg Message) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	log.Printf("Broadcasting message from %s to %d clients", msg.From, len(r.clients))
	for nickname, client := range r.clients {
		log.Printf("Sending to client: %s", nickname)
		go client.sendMessage(msg) // Use goroutine to avoid blocking
	}
}

// Join adds a client to the room
func (r *Room) Join(client *Client) {
	r.join <- client
}

// Leave removes a client from the room
func (r *Room) Leave(client *Client) {
	r.leave <- client
}

// Broadcast sends a message to all clients
func (r *Room) Broadcast(msg Message) {
	r.broadcast <- msg
}

// GetUserList returns a list of all users in the room
func (r *Room) GetUserList() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	users := make([]string, 0, len(r.clients))
	for nickname := range r.clients {
		users = append(users, nickname)
	}
	
	return users
}

// IsNicknameAvailable checks if a nickname is available
func (r *Room) IsNicknameAvailable(nickname string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	_, exists := r.clients[nickname]
	return !exists
}
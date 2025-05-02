package chat

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/your-username/ts-chat/internal/ui"
)

// Client represents a chat client
type Client struct {
	Nickname string
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
	room     *Room
	mu       sync.Mutex // Mutex to protect concurrent writes
}

// NewClient creates a new chat client
func NewClient(conn net.Conn, room *Room) (*Client, error) {
	client := &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
		room:   room,
	}
	
	// Ask for nickname
	if err := client.requestNickname(); err != nil {
		return nil, err
	}
	
	// Join the room
	room.Join(client)
	
	// Send welcome message
	client.sendWelcomeMessage()
	
	return client, nil
}

// requestNickname asks the user for a nickname
func (c *Client) requestNickname() error {
	// Send welcome message
	c.write(ui.FormatTitle("Welcome to Tailscale Terminal Chat") + "\r\n\r\n")
	
	// Ask for nickname
	for {
		c.write(ui.InputStyle.Render("Please enter your nickname: "))
		
		// Read nickname
		nickname, err := c.reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read nickname: %w", err)
		}
		
		// Trim whitespace
		nickname = strings.TrimSpace(nickname)
		
		// Validate nickname
		if nickname == "" {
			c.write("Nickname cannot be empty. Please try again.\r\n")
			continue
		}
		
		if strings.ToLower(nickname) == "system" {
			c.write("Nickname 'System' is reserved. Please choose another nickname.\r\n")
			continue
		}
		
		if !c.room.IsNicknameAvailable(nickname) {
			c.write(fmt.Sprintf("Nickname '%s' is already taken. Please choose another nickname.\r\n", nickname))
			continue
		}
		
		// Set nickname
		c.Nickname = nickname
		break
	}
	
	return nil
}

// sendWelcomeMessage sends a welcome message to the client
func (c *Client) sendWelcomeMessage() {
	banner := `
╔═══════════════════════════════════════════════════════════════════════╗
║           _____                    _             _   _____             ║
║          |_   _|__ _ __ _ __ ___ (_)_ __   __ _| | |  __ \            ║
║            | |/ _ \ '__| '_ ' _ \| | '_ \ / _' | | | |  | |           ║
║            | |  __/ |  | | | | | | | | | | (_| | | | |__| |           ║
║            |_|\___|_|  |_| |_| |_|_|_| |_|\__,_|_| |_____/            ║
║                                                                       ║
║                             CHAT ROOM                                 ║
╚═══════════════════════════════════════════════════════════════════════╝
`
	coloredBanner := ui.SystemStyle.Render(banner)
	welcomeMsg := ui.FormatWelcomeMessage(c.room.Name, c.Nickname)
	
	c.write(coloredBanner + "\r\n")
	c.write(welcomeMsg + "\r\n\r\n")
	c.write("Type a message and press Enter to send. Type /help for commands.\r\n\r\n")
}

// Handle handles client interactions
func (c *Client) Handle(ctx context.Context) {
	// Cleanup when done
	defer c.room.Leave(c)
	
	// Handle client messages
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Read message from client
			line, err := c.reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					// Client disconnected
					return
				}
				c.sendSystemMessage(fmt.Sprintf("Error reading message: %v", err))
				return
			}
			
			// Trim whitespace
			message := strings.TrimSpace(line)
			
			// Handle command or message
			if strings.HasPrefix(message, "/") {
				c.handleCommand(message)
			} else if message != "" {
				// Send message to room
				c.room.Broadcast(Message{
					From:      c.Nickname,
					Content:   message,
					Timestamp: time.Now(),
				})
			}
		}
	}
}

// handleCommand handles a command from the client
func (c *Client) handleCommand(cmd string) {
	parts := strings.SplitN(cmd, " ", 2)
	command := strings.ToLower(parts[0])
	
	switch command {
	case "/who":
		c.showUserList()
	case "/me":
		if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
			c.sendSystemMessage("Usage: /me <action>")
			return
		}
		action := parts[1]
		c.room.Broadcast(Message{
			From:      c.Nickname,
			Content:   action,
			Timestamp: time.Now(),
			IsAction:  true,
		})
	case "/help":
		c.showHelp()
	case "/quit":
		c.sendSystemMessage("Goodbye!")
		c.conn.Close()
	default:
		c.sendSystemMessage(fmt.Sprintf("Unknown command: %s", command))
	}
}

// showUserList shows the list of users in the room
func (c *Client) showUserList() {
	users := c.room.GetUserList()
	msg := ui.FormatUserList(c.room.Name, users, c.room.MaxUsers)
	c.write(msg + "\r\n")
}

// showHelp shows the help message
func (c *Client) showHelp() {
	helpMsg := ui.FormatHelp()
	c.write(helpMsg + "\r\n")
}

// sendSystemMessage sends a system message to the client
func (c *Client) sendSystemMessage(message string) {
	msg := Message{
		From:      "System",
		Content:   message,
		Timestamp: time.Now(),
		IsSystem:  true,
	}
	
	c.sendMessage(msg)
}

// sendMessage sends a message to the client
func (c *Client) sendMessage(msg Message) {
	var formatted string
	timeStr := msg.Timestamp.Format("15:04:05")
	
	// Log the message for debugging
	log.Printf("Sending message from %s to %s: %s", msg.From, c.Nickname, msg.Content)
	
	if msg.IsSystem {
		formatted = ui.FormatSystemMessage(msg.Content) + "\r\n"
	} else if msg.IsAction {
		formatted = ui.FormatActionMessage(msg.From, msg.Content) + "\r\n"
	} else if msg.From == c.Nickname {
		formatted = ui.FormatSelfMessage(msg.Content, timeStr) + "\r\n"
	} else {
		formatted = ui.FormatUserMessage(msg.From, msg.Content, timeStr) + "\r\n"
	}
	
	// Ensure the write happens without blocking
	go func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.writer.WriteString(formatted)
		c.writer.Flush()
	}()
}

// write writes a message to the client
func (c *Client) write(message string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writer.WriteString(message)
	c.writer.Flush()
}
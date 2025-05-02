package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/your-username/ts-chat/internal/chat"
	"tailscale.com/tsnet"
)

// Server represents the chat server
type Server struct {
	config      Config
	listener    net.Listener
	tsServer    *tsnet.Server
	chatRoom    *chat.Room
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	connections map[string]net.Conn
	mu          sync.Mutex
}

// NewServer creates a new chat server
func NewServer(cfg Config) (*Server, error) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create a new chat room
	room := chat.NewRoom(cfg.RoomName, cfg.MaxUsers)
	
	return &Server{
		config:      cfg,
		ctx:         ctx,
		cancel:      cancel,
		chatRoom:    room,
		connections: make(map[string]net.Conn),
	}, nil
}

// Start starts the chat server
func (s *Server) Start() error {
	var listener net.Listener
	var err error
	
	if s.config.EnableTailscale {
		// Start the tsnet Tailscale server
		s.tsServer = &tsnet.Server{
			Hostname: s.config.HostName,
			AuthKey:  os.Getenv("TS_AUTHKEY"),
		}
		
		// Listen on the specified port
		listener, err = s.tsServer.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
		if err != nil {
			return fmt.Errorf("failed to start Tailscale server on port %d: %w", s.config.Port, err)
		}
		
		// Try to get Tailscale status
		ln, err := s.tsServer.LocalClient()
		if err != nil {
			log.Printf("Warning: unable to get Tailscale local client: %v", err)
		} else {
			status, err := ln.Status(s.ctx)
			if err != nil {
				log.Printf("Warning: unable to get Tailscale status: %v", err)
			} else if status.Self.DNSName != "" {
				log.Printf("Tailscale node running as: %s", status.Self.DNSName)
			}
		}
	} else {
		// Start a regular TCP server
		listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
		if err != nil {
			return fmt.Errorf("failed to listen on port %d: %w", s.config.Port, err)
		}
	}
	
	s.listener = listener
	
	log.Printf("Server started on port %d", s.config.Port)
	log.Printf("Room name: %s", s.config.RoomName)
	log.Printf("Maximum users: %d", s.config.MaxUsers)
	
	// Accept connections
	s.wg.Add(1)
	go s.acceptConnections()
	
	return nil
}

// acceptConnections accepts incoming connections
func (s *Server) acceptConnections() {
	defer s.wg.Done()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				// Check if server is shutting down
				select {
				case <-s.ctx.Done():
					return
				default:
					log.Printf("Error accepting connection: %v", err)
					continue
				}
			}
			
			// Handle the connection in a new goroutine
			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}
}

// handleConnection handles a client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()
	
	remoteAddr := conn.RemoteAddr().String()
	log.Printf("New connection from %s", remoteAddr)
	
	// Register connection
	s.mu.Lock()
	s.connections[remoteAddr] = conn
	s.mu.Unlock()
	
	// Deregister connection when done
	defer func() {
		s.mu.Lock()
		delete(s.connections, remoteAddr)
		s.mu.Unlock()
		log.Printf("Connection from %s closed", remoteAddr)
	}()
	
	// Create a new client
	client, err := chat.NewClient(conn, s.chatRoom)
	if err != nil {
		log.Printf("Error creating client: %v", err)
		return
	}
	
	// Handle the client
	client.Handle(s.ctx)
}

// Stop stops the chat server
func (s *Server) Stop() error {
	// Cancel the context to signal shutdown
	s.cancel()
	
	// Close all active connections
	s.mu.Lock()
	for addr, conn := range s.connections {
		log.Printf("Closing connection from %s", addr)
		conn.Close()
	}
	s.mu.Unlock()
	
	// Close the listener
	if s.listener != nil {
		log.Print("Closing listener")
		if err := s.listener.Close(); err != nil {
			log.Printf("Error closing listener: %v", err)
		}
	}
	
	// Close the tsnet server if in Tailscale mode
	if s.config.EnableTailscale && s.tsServer != nil {
		log.Print("Closing Tailscale node")
		if err := s.tsServer.Close(); err != nil {
			log.Printf("Error closing Tailscale node: %v", err)
		}
	}
	
	// Wait for all goroutines to finish
	s.wg.Wait()
	
	return nil
}
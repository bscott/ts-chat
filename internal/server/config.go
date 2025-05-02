package server

// Config holds the server configuration
type Config struct {
	Port           int    // TCP port to listen on
	RoomName       string // Chat room name
	MaxUsers       int    // Maximum allowed users
	EnableTailscale bool   // Whether to enable Tailscale mode
	HostName       string // Tailscale hostname (only used if EnableTailscale is true)
}
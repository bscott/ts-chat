# Terminal Chat with Tailscale

A terminal-based chat application built in Go. The chat server allows multiple users to connect via telnet or netcat and chat in real-time. It supports both regular TCP mode and Tailscale mode for secure networking.

## Features

- Terminal-based interface with styled text using ANSI colors
- Basic chat commands: `/who`, `/me`, `/help`, `/quit`
- Configurable port, room name, and maximum number of users
- Optional Tailscale integration for secure networking across devices
- Support for simultaneous connections

## Requirements

- Go 1.19+
- Tailscale account (only required for Tailscale mode)

## Installation

```bash
# Clone the repository
git clone https://github.com/your-username/ts-chat.git
cd ts-chat

# Build the binary
go build -o chat-server ./cmd/ts-chat

# Or use the provided Makefile
make
```

## Usage

### Regular Mode:

Run the server in regular TCP mode (accessible only on the local network):

```bash
# Run with default settings
./chat-server

# Run with custom settings
./chat-server --port 2323 --room-name "My Chat Room" --max-users 20
```

### Tailscale Mode:

Run the server in Tailscale mode (accessible over your Tailnet):

```bash
# Set your Tailscale auth key
export TS_AUTHKEY=tskey-your-auth-key-here

# Run with Tailscale mode enabled
./chat-server --tailscale --hostname mychat --room-name "Tailscale Chat" --port 2323
```

When using Tailscale mode, the chat server will:
1. Authenticate with Tailscale using your TS_AUTHKEY
2. Register a node in your Tailnet with the specified hostname
3. Be accessible from any device on your Tailnet

### Configuration options:

- `--port`: TCP port to listen on (default: 2323)
- `--room-name`: Chat room name (default: "Chat Room")
- `--max-users`: Maximum allowed users (default: 10)
- `--tailscale`: Enable Tailscale mode (default: false)
- `--hostname`: Tailscale hostname (default: "chatroom", only used if --tailscale is enabled)

### Tailscale Authentication:

To use Tailscale mode, you need to provide an auth key:

1. Obtain a Tailscale auth key from the [Tailscale Admin Console](https://login.tailscale.com/admin/settings/keys)
2. Set the auth key as an environment variable:
   ```bash
   export TS_AUTHKEY=tskey-your-auth-key-here
   ```

### Docker usage:

```bash
# Build the Docker image
docker build -t chat-server .

# Run in regular mode
docker run -p 2323:2323 chat-server

# Run in Tailscale mode
docker run -e TS_AUTHKEY=tskey-your-auth-key-here chat-server --tailscale --hostname dockerchat
```

### Connecting to the chat:

#### Regular mode:
```bash
# Connect via Netcat
nc localhost 2323

# Or Telnet
telnet localhost 2323
```

#### Tailscale mode:
```bash
# Connect via Netcat (replace 'hostname' with your specified hostname)
nc hostname.ts.net 2323

# Or Telnet
telnet hostname.ts.net 2323
```

## Chat Commands

When connected to the chat, the following commands are available:

- `/who` - Shows a list of all users in the room
- `/me <action>` - Perform an action (e.g., `/me waves hello` displays `* Username waves hello`)
- `/help` - Shows the available commands
- `/quit` - Disconnects from the chat

## Development

The project is organized as follows:

- `cmd/ts-chat/main.go`: Main application entry point
- `internal/server/`: Server implementation
- `internal/chat/`: Chat room and client handling
- `internal/ui/`: Terminal UI styling

## License

MIT
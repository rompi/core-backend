# Package Plan: pkg/websocket

## Overview

A real-time WebSocket server package with support for rooms/channels, pub/sub messaging, connection management, and horizontal scaling via Redis. Designed for building chat applications, live notifications, real-time dashboards, and collaborative features.

## Goals

1. **Connection Management** - Handle thousands of concurrent connections
2. **Rooms/Channels** - Group connections for targeted messaging
3. **Pub/Sub** - Publish messages to subscribers
4. **Authentication** - Integrate with pkg/auth for connection auth
5. **Horizontal Scaling** - Redis adapter for multi-instance deployments
6. **Binary & Text** - Support both message formats
7. **Heartbeat/Ping** - Connection health monitoring

## Architecture

```
pkg/websocket/
├── websocket.go          # Core Server interface
├── config.go             # Configuration
├── options.go            # Functional options
├── connection.go         # Connection management
├── room.go               # Room/channel management
├── message.go            # Message types
├── hub.go                # Connection hub
├── errors.go             # Custom error types
├── adapter/
│   ├── adapter.go        # Adapter interface
│   ├── memory.go         # In-memory adapter (single instance)
│   └── redis.go          # Redis adapter (multi-instance)
├── middleware/
│   ├── auth.go           # Authentication middleware
│   ├── ratelimit.go      # Rate limiting
│   └── logging.go        # Connection logging
├── examples/
│   ├── basic/
│   ├── chat/
│   ├── notifications/
│   └── with-redis/
└── README.md
```

## Core Interfaces

```go
package websocket

import (
    "context"
    "net/http"
    "time"
)

// Server manages WebSocket connections
type Server interface {
    // Handler returns HTTP handler for upgrades
    Handler() http.Handler

    // OnConnect registers connection callback
    OnConnect(fn func(conn *Conn))

    // OnDisconnect registers disconnection callback
    OnDisconnect(fn func(conn *Conn))

    // OnMessage registers message handler
    OnMessage(event string, handler MessageHandler)

    // Broadcast sends to all connections
    Broadcast(ctx context.Context, event string, data interface{}) error

    // BroadcastToRoom sends to room members
    BroadcastToRoom(ctx context.Context, room, event string, data interface{}) error

    // Send sends to specific connection
    Send(ctx context.Context, connID, event string, data interface{}) error

    // GetConnection returns connection by ID
    GetConnection(id string) (*Conn, bool)

    // GetConnections returns all connections
    GetConnections() []*Conn

    // GetRoomConnections returns connections in a room
    GetRoomConnections(room string) []*Conn

    // Stats returns server statistics
    Stats() Stats

    // Close shuts down the server
    Close() error
}

// Conn represents a WebSocket connection
type Conn struct {
    // ID is the unique connection identifier
    ID string

    // UserID is the authenticated user (if any)
    UserID string

    // Metadata holds custom connection data
    Metadata map[string]interface{}

    // Rooms the connection has joined
    Rooms []string

    // RemoteAddr is the client address
    RemoteAddr string

    // ConnectedAt is connection time
    ConnectedAt time.Time
}

// Methods
func (c *Conn) Send(event string, data interface{}) error
func (c *Conn) Join(room string) error
func (c *Conn) Leave(room string) error
func (c *Conn) LeaveAll() error
func (c *Conn) Close() error
func (c *Conn) SetMetadata(key string, value interface{})
func (c *Conn) GetMetadata(key string) interface{}

// MessageHandler handles incoming messages
type MessageHandler func(ctx context.Context, conn *Conn, msg *Message)

// Message represents a WebSocket message
type Message struct {
    // Event is the message type/event name
    Event string `json:"event"`

    // Data is the message payload
    Data json.RawMessage `json:"data"`

    // Room for room-specific messages
    Room string `json:"room,omitempty"`

    // To for direct messages
    To string `json:"to,omitempty"`
}

// Stats holds server statistics
type Stats struct {
    Connections   int64
    Rooms         int64
    MessagesSent  int64
    MessagesRecv  int64
    BytesSent     int64
    BytesRecv     int64
}
```

## Configuration

```go
// Config holds WebSocket server configuration
type Config struct {
    // Read buffer size
    ReadBufferSize int `env:"WS_READ_BUFFER_SIZE" default:"1024"`

    // Write buffer size
    WriteBufferSize int `env:"WS_WRITE_BUFFER_SIZE" default:"1024"`

    // Maximum message size
    MaxMessageSize int64 `env:"WS_MAX_MESSAGE_SIZE" default:"65536"`

    // Write timeout
    WriteTimeout time.Duration `env:"WS_WRITE_TIMEOUT" default:"10s"`

    // Ping interval
    PingInterval time.Duration `env:"WS_PING_INTERVAL" default:"30s"`

    // Pong timeout
    PongTimeout time.Duration `env:"WS_PONG_TIMEOUT" default:"60s"`

    // Enable compression
    EnableCompression bool `env:"WS_ENABLE_COMPRESSION" default:"true"`

    // Allowed origins (empty = allow all)
    AllowedOrigins []string `env:"WS_ALLOWED_ORIGINS"`

    // Adapter type: "memory" or "redis"
    Adapter string `env:"WS_ADAPTER" default:"memory"`
}

// RedisAdapterConfig for horizontal scaling
type RedisAdapterConfig struct {
    // Redis URL
    URL string `env:"WS_REDIS_URL" default:"redis://localhost:6379"`

    // Channel prefix
    Prefix string `env:"WS_REDIS_PREFIX" default:"ws:"`

    // Pool size
    PoolSize int `env:"WS_REDIS_POOL_SIZE" default:"10"`
}
```

## Usage Examples

### Basic Server

```go
package main

import (
    "context"
    "log"
    "net/http"
    "github.com/user/core-backend/pkg/websocket"
)

func main() {
    // Create WebSocket server
    ws := websocket.New(websocket.Config{
        PingInterval: 30 * time.Second,
    })

    // Handle connections
    ws.OnConnect(func(conn *websocket.Conn) {
        log.Printf("Client connected: %s", conn.ID)
    })

    ws.OnDisconnect(func(conn *websocket.Conn) {
        log.Printf("Client disconnected: %s", conn.ID)
    })

    // Handle messages
    ws.OnMessage("chat", func(ctx context.Context, conn *websocket.Conn, msg *websocket.Message) {
        var chatMsg ChatMessage
        json.Unmarshal(msg.Data, &chatMsg)

        // Broadcast to room
        ws.BroadcastToRoom(ctx, chatMsg.Room, "chat", chatMsg)
    })

    // Mount handler
    http.Handle("/ws", ws.Handler())
    http.ListenAndServe(":8080", nil)
}
```

### Chat Application

```go
type ChatMessage struct {
    Room    string `json:"room"`
    User    string `json:"user"`
    Content string `json:"content"`
}

func main() {
    ws := websocket.New(cfg)

    // Join room
    ws.OnMessage("join", func(ctx context.Context, conn *websocket.Conn, msg *websocket.Message) {
        var req struct {
            Room string `json:"room"`
        }
        json.Unmarshal(msg.Data, &req)

        conn.Join(req.Room)
        conn.Send("joined", map[string]string{"room": req.Room})

        // Notify others
        ws.BroadcastToRoom(ctx, req.Room, "user_joined", map[string]string{
            "user": conn.UserID,
        })
    })

    // Leave room
    ws.OnMessage("leave", func(ctx context.Context, conn *websocket.Conn, msg *websocket.Message) {
        var req struct {
            Room string `json:"room"`
        }
        json.Unmarshal(msg.Data, &req)

        conn.Leave(req.Room)

        ws.BroadcastToRoom(ctx, req.Room, "user_left", map[string]string{
            "user": conn.UserID,
        })
    })

    // Chat message
    ws.OnMessage("message", func(ctx context.Context, conn *websocket.Conn, msg *websocket.Message) {
        var chatMsg ChatMessage
        json.Unmarshal(msg.Data, &chatMsg)

        chatMsg.User = conn.UserID

        ws.BroadcastToRoom(ctx, chatMsg.Room, "message", chatMsg)
    })
}
```

### With Authentication

```go
import (
    "github.com/user/core-backend/pkg/websocket"
    "github.com/user/core-backend/pkg/websocket/middleware"
    "github.com/user/core-backend/pkg/auth"
)

func main() {
    authService := auth.New(/*...*/)

    ws := websocket.New(cfg,
        websocket.WithMiddleware(
            middleware.Auth(authService, middleware.AuthConfig{
                // Extract token from query param or header
                TokenExtractor: func(r *http.Request) string {
                    if token := r.URL.Query().Get("token"); token != "" {
                        return token
                    }
                    return r.Header.Get("Authorization")
                },
            }),
        ),
    )

    ws.OnConnect(func(conn *websocket.Conn) {
        // conn.UserID is set by auth middleware
        log.Printf("User %s connected", conn.UserID)
    })
}
```

### With Redis Adapter (Horizontal Scaling)

```go
import (
    "github.com/user/core-backend/pkg/websocket"
    "github.com/user/core-backend/pkg/websocket/adapter"
)

func main() {
    redisAdapter, err := adapter.NewRedis(adapter.RedisConfig{
        URL:    "redis://localhost:6379",
        Prefix: "myapp:ws:",
    })
    if err != nil {
        log.Fatal(err)
    }

    ws := websocket.New(cfg,
        websocket.WithAdapter(redisAdapter),
    )

    // Broadcasts are now distributed across all instances
    ws.BroadcastToRoom(ctx, "general", "message", data)
}
```

### Direct Messaging

```go
ws.OnMessage("dm", func(ctx context.Context, conn *websocket.Conn, msg *websocket.Message) {
    var dm struct {
        To      string `json:"to"`
        Content string `json:"content"`
    }
    json.Unmarshal(msg.Data, &dm)

    // Send directly to user
    ws.Send(ctx, dm.To, "dm", map[string]interface{}{
        "from":    conn.UserID,
        "content": dm.Content,
    })
})
```

### Client JavaScript Example

```javascript
const ws = new WebSocket('ws://localhost:8080/ws?token=jwt-token');

ws.onopen = () => {
    // Join a room
    ws.send(JSON.stringify({
        event: 'join',
        data: { room: 'general' }
    }));
};

ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);

    switch (msg.event) {
        case 'message':
            console.log('New message:', msg.data);
            break;
        case 'user_joined':
            console.log('User joined:', msg.data.user);
            break;
    }
};

// Send a message
ws.send(JSON.stringify({
    event: 'message',
    data: {
        room: 'general',
        content: 'Hello, world!'
    }
}));
```

## Dependencies

- **Required:** `github.com/gorilla/websocket`
- **Optional:** `github.com/redis/go-redis/v9` for Redis adapter

## Implementation Phases

### Phase 1: Core Server
1. WebSocket upgrade handling
2. Connection management
3. Message handling
4. Ping/pong heartbeat

### Phase 2: Rooms & Broadcasting
1. Room join/leave
2. Broadcast to all/room
3. Direct messaging

### Phase 3: Adapters
1. Memory adapter
2. Redis adapter for scaling

### Phase 4: Middleware & Auth
1. Authentication middleware
2. Rate limiting
3. Logging

### Phase 5: Documentation & Examples
1. README
2. Chat example
3. Notifications example

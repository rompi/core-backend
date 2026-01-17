# Package Plan: pkg/notification

## Overview

A push notification package supporting multiple providers (Firebase Cloud Messaging, Apple Push Notification Service, Web Push). Provides a unified API for sending notifications to mobile and web clients with support for topics, targeting, and delivery tracking.

## Goals

1. **Multiple Providers** - FCM, APNs, Web Push
2. **Unified Interface** - Single API for all platforms
3. **Batch Sending** - Efficient bulk notifications
4. **Topics/Channels** - Subscribe users to topics
5. **Scheduling** - Schedule notifications for later
6. **Templates** - Reusable notification templates
7. **Delivery Tracking** - Track delivery status

## Architecture

```
pkg/notification/
├── notification.go       # Core interfaces
├── config.go             # Configuration
├── options.go            # Functional options
├── message.go            # Message types
├── errors.go             # Custom error types
├── provider/
│   ├── provider.go       # Provider interface
│   ├── fcm.go            # Firebase Cloud Messaging
│   ├── apns.go           # Apple Push Notification Service
│   └── webpush.go        # Web Push (VAPID)
├── examples/
│   ├── basic/
│   ├── fcm/
│   ├── apns/
│   └── multiplatform/
└── README.md
```

## Core Interfaces

```go
package notification

import (
    "context"
    "time"
)

// Sender sends push notifications
type Sender interface {
    // Send sends a notification to specific tokens
    Send(ctx context.Context, msg *Message, tokens ...string) (*Response, error)

    // SendToTopic sends to a topic
    SendToTopic(ctx context.Context, msg *Message, topic string) (*Response, error)

    // SendMulticast sends to multiple tokens
    SendMulticast(ctx context.Context, msg *Message, tokens []string) (*BatchResponse, error)

    // SubscribeToTopic subscribes tokens to a topic
    SubscribeToTopic(ctx context.Context, topic string, tokens ...string) error

    // UnsubscribeFromTopic unsubscribes tokens from a topic
    UnsubscribeFromTopic(ctx context.Context, topic string, tokens ...string) error

    // Close releases resources
    Close() error
}

// Message represents a notification
type Message struct {
    // Title is the notification title
    Title string `json:"title"`

    // Body is the notification body
    Body string `json:"body"`

    // Image URL for rich notifications
    ImageURL string `json:"image_url,omitempty"`

    // Data payload (key-value pairs)
    Data map[string]string `json:"data,omitempty"`

    // Android-specific options
    Android *AndroidConfig `json:"android,omitempty"`

    // iOS-specific options
    iOS *IOSConfig `json:"ios,omitempty"`

    // Web-specific options
    Web *WebConfig `json:"web,omitempty"`

    // Priority: "high" or "normal"
    Priority string `json:"priority,omitempty"`

    // TTL is time-to-live
    TTL time.Duration `json:"ttl,omitempty"`

    // CollapseKey for collapsing notifications
    CollapseKey string `json:"collapse_key,omitempty"`
}

// AndroidConfig for Android-specific options
type AndroidConfig struct {
    ChannelID    string `json:"channel_id,omitempty"`
    Icon         string `json:"icon,omitempty"`
    Color        string `json:"color,omitempty"`
    Sound        string `json:"sound,omitempty"`
    Tag          string `json:"tag,omitempty"`
    ClickAction  string `json:"click_action,omitempty"`
}

// IOSConfig for iOS-specific options
type IOSConfig struct {
    Badge        *int   `json:"badge,omitempty"`
    Sound        string `json:"sound,omitempty"`
    Category     string `json:"category,omitempty"`
    ThreadID     string `json:"thread_id,omitempty"`
    MutableContent bool `json:"mutable_content,omitempty"`
}

// WebConfig for Web Push options
type WebConfig struct {
    Icon     string            `json:"icon,omitempty"`
    Badge    string            `json:"badge,omitempty"`
    Actions  []WebAction       `json:"actions,omitempty"`
    Vibrate  []int             `json:"vibrate,omitempty"`
    Data     map[string]string `json:"data,omitempty"`
}

// WebAction for web notification actions
type WebAction struct {
    Action string `json:"action"`
    Title  string `json:"title"`
    Icon   string `json:"icon,omitempty"`
}

// Response for single notification
type Response struct {
    MessageID string
    Success   bool
    Error     error
}

// BatchResponse for multicast
type BatchResponse struct {
    SuccessCount int
    FailureCount int
    Responses    []Response
}
```

## Configuration

```go
// Config holds notification configuration
type Config struct {
    // Provider: "fcm", "apns", "webpush", "multi"
    Provider string `env:"NOTIFICATION_PROVIDER" default:"fcm"`
}

// FCMConfig for Firebase Cloud Messaging
type FCMConfig struct {
    // Credentials JSON file path
    CredentialsFile string `env:"FCM_CREDENTIALS_FILE"`

    // Credentials JSON string
    CredentialsJSON string `env:"FCM_CREDENTIALS_JSON"`

    // Project ID
    ProjectID string `env:"FCM_PROJECT_ID"`
}

// APNsConfig for Apple Push Notification Service
type APNsConfig struct {
    // Certificate file path (.p12)
    CertFile string `env:"APNS_CERT_FILE"`

    // Certificate password
    CertPassword string `env:"APNS_CERT_PASSWORD"`

    // Auth key file path (.p8)
    AuthKeyFile string `env:"APNS_AUTH_KEY_FILE"`

    // Key ID
    KeyID string `env:"APNS_KEY_ID"`

    // Team ID
    TeamID string `env:"APNS_TEAM_ID"`

    // Bundle ID
    BundleID string `env:"APNS_BUNDLE_ID"`

    // Use production environment
    Production bool `env:"APNS_PRODUCTION" default:"false"`
}

// WebPushConfig for Web Push
type WebPushConfig struct {
    // VAPID public key
    VAPIDPublicKey string `env:"WEBPUSH_VAPID_PUBLIC_KEY" required:"true"`

    // VAPID private key
    VAPIDPrivateKey string `env:"WEBPUSH_VAPID_PRIVATE_KEY" required:"true"`

    // Contact email
    VAPIDContact string `env:"WEBPUSH_VAPID_CONTACT" required:"true"`
}
```

## Usage Examples

### Basic FCM

```go
package main

import (
    "context"
    "github.com/user/core-backend/pkg/notification"
    "github.com/user/core-backend/pkg/notification/provider"
)

func main() {
    fcm, _ := provider.NewFCM(provider.FCMConfig{
        CredentialsFile: "firebase-credentials.json",
    })
    defer fcm.Close()

    ctx := context.Background()

    msg := &notification.Message{
        Title: "New Message",
        Body:  "You have a new message from John",
        Data: map[string]string{
            "type":    "message",
            "chat_id": "123",
        },
        Android: &notification.AndroidConfig{
            ChannelID: "messages",
            Icon:      "ic_notification",
        },
    }

    // Send to device token
    resp, _ := fcm.Send(ctx, msg, "device-token-here")
    fmt.Printf("Message ID: %s\n", resp.MessageID)

    // Send to topic
    fcm.SendToTopic(ctx, msg, "news")
}
```

### Multi-Platform

```go
func main() {
    // Create multi-provider sender
    sender := notification.NewMulti(
        provider.NewFCM(fcmConfig),   // Android
        provider.NewAPNs(apnsConfig), // iOS
        provider.NewWebPush(webConfig), // Web
    )

    msg := &notification.Message{
        Title: "Order Shipped",
        Body:  "Your order #12345 has been shipped!",
        Data: map[string]string{
            "order_id": "12345",
        },
    }

    // Send to all platforms based on token prefix/format
    sender.Send(ctx, msg,
        "fcm:android-token",
        "apns:ios-device-token",
        "web:webpush-subscription-json",
    )
}
```

### Topic Subscriptions

```go
func main() {
    fcm, _ := provider.NewFCM(cfg)

    // Subscribe user to topics
    fcm.SubscribeToTopic(ctx, "news", userToken)
    fcm.SubscribeToTopic(ctx, "deals", userToken)

    // Send to all news subscribers
    fcm.SendToTopic(ctx, &notification.Message{
        Title: "Breaking News",
        Body:  "Important update...",
    }, "news")
}
```

### Batch Sending

```go
func main() {
    fcm, _ := provider.NewFCM(cfg)

    tokens := []string{"token1", "token2", "token3", /* ... */}

    resp, _ := fcm.SendMulticast(ctx, msg, tokens)

    fmt.Printf("Success: %d, Failed: %d\n", resp.SuccessCount, resp.FailureCount)

    // Handle failures
    for i, r := range resp.Responses {
        if !r.Success {
            fmt.Printf("Token %s failed: %v\n", tokens[i], r.Error)
        }
    }
}
```

## Dependencies

- **Required:** None (interface only)
- **Optional:**
  - `firebase.google.com/go/v4/messaging` for FCM
  - `github.com/sideshow/apns2` for APNs
  - `github.com/SherClockHolmes/webpush-go` for Web Push

## Implementation Phases

### Phase 1: Core Interface & FCM
1. Define Sender interface
2. Message types
3. FCM provider

### Phase 2: APNs Provider
1. Certificate auth
2. Token auth
3. iOS-specific options

### Phase 3: Web Push Provider
1. VAPID authentication
2. Subscription handling

### Phase 4: Advanced Features
1. Multi-provider sender
2. Topic management
3. Batch sending

### Phase 5: Documentation
1. README
2. Platform-specific examples

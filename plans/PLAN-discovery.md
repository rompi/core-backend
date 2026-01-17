# Package Plan: pkg/discovery

## Overview

A service discovery package for dynamic service location in microservices architectures. Supports multiple backends (Consul, etcd, Kubernetes, DNS) with features like health checking, load balancing, and watch capabilities.

## Goals

1. **Multiple Backends** - Consul, etcd, Kubernetes, DNS, static
2. **Service Registration** - Register services with metadata
3. **Health Integration** - Automatic health-based deregistration
4. **Watch Support** - Real-time service change notifications
5. **Load Balancing** - Client-side load balancing strategies
6. **Caching** - Local caching with TTL
7. **Graceful Shutdown** - Clean deregistration on shutdown

## Architecture

```
pkg/discovery/
├── discovery.go          # Core interfaces
├── config.go             # Configuration
├── options.go            # Functional options
├── service.go            # Service definition
├── resolver.go           # Service resolver
├── provider/
│   ├── provider.go       # Provider interface
│   ├── consul.go         # Consul provider
│   ├── etcd.go           # etcd provider
│   ├── kubernetes.go     # Kubernetes provider
│   ├── dns.go            # DNS-based discovery
│   └── static.go         # Static configuration
├── balancer/
│   ├── balancer.go       # Load balancer interface
│   ├── roundrobin.go     # Round-robin
│   ├── random.go         # Random selection
│   ├── weightedrr.go     # Weighted round-robin
│   └── leastconn.go      # Least connections
├── examples/
│   ├── basic/
│   ├── consul/
│   └── kubernetes/
└── README.md
```

## Core Interfaces

```go
package discovery

import (
    "context"
    "time"
)

// Discovery manages service discovery
type Discovery interface {
    // Register registers a service
    Register(ctx context.Context, service *Service) error

    // Deregister removes a service
    Deregister(ctx context.Context, serviceID string) error

    // Discover finds service instances
    Discover(ctx context.Context, name string) ([]*Instance, error)

    // Watch watches for service changes
    Watch(ctx context.Context, name string, callback func([]*Instance)) error

    // Resolve returns a single instance (with load balancing)
    Resolve(ctx context.Context, name string) (*Instance, error)

    // Close releases resources
    Close() error
}

// Service represents a service to register
type Service struct {
    ID       string
    Name     string
    Version  string
    Address  string
    Port     int
    Tags     []string
    Metadata map[string]string
    Health   *HealthCheck
}

// Instance represents a discovered service instance
type Instance struct {
    ID        string
    Name      string
    Version   string
    Address   string
    Port      int
    Tags      []string
    Metadata  map[string]string
    Healthy   bool
    Weight    int
}

// Endpoint returns the instance endpoint
func (i *Instance) Endpoint() string {
    return fmt.Sprintf("%s:%d", i.Address, i.Port)
}

// HealthCheck defines health check configuration
type HealthCheck struct {
    HTTP     string        // HTTP URL to check
    GRPC     string        // gRPC service to check
    TCP      string        // TCP address to check
    Interval time.Duration // Check interval
    Timeout  time.Duration // Check timeout
}

// Balancer selects instances
type Balancer interface {
    // Pick selects an instance
    Pick(instances []*Instance) *Instance

    // Name returns the balancer name
    Name() string
}
```

## Configuration

```go
// Config holds discovery configuration
type Config struct {
    // Provider: "consul", "etcd", "kubernetes", "dns", "static"
    Provider string `env:"DISCOVERY_PROVIDER" default:"consul"`

    // Service name for registration
    ServiceName string `env:"DISCOVERY_SERVICE_NAME"`

    // Load balancing strategy
    Balancer string `env:"DISCOVERY_BALANCER" default:"round_robin"`

    // Cache TTL
    CacheTTL time.Duration `env:"DISCOVERY_CACHE_TTL" default:"10s"`

    // Enable health checks
    HealthCheck bool `env:"DISCOVERY_HEALTH_CHECK" default:"true"`
}

// ConsulConfig for Consul provider
type ConsulConfig struct {
    Address    string `env:"CONSUL_HTTP_ADDR" default:"localhost:8500"`
    Token      string `env:"CONSUL_HTTP_TOKEN"`
    Datacenter string `env:"CONSUL_DATACENTER"`
    Namespace  string `env:"CONSUL_NAMESPACE"`
}

// EtcdConfig for etcd provider
type EtcdConfig struct {
    Endpoints []string `env:"ETCD_ENDPOINTS" default:"localhost:2379"`
    Prefix    string   `env:"ETCD_DISCOVERY_PREFIX" default:"/services/"`
    Username  string   `env:"ETCD_USERNAME"`
    Password  string   `env:"ETCD_PASSWORD"`
}

// KubernetesConfig for Kubernetes provider
type KubernetesConfig struct {
    Namespace    string `env:"K8S_NAMESPACE" default:"default"`
    LabelSelector string `env:"K8S_LABEL_SELECTOR"`
    InCluster    bool   `env:"K8S_IN_CLUSTER" default:"true"`
}

// DNSConfig for DNS provider
type DNSConfig struct {
    Server     string        `env:"DNS_SERVER" default:""`
    Port       int           `env:"DNS_PORT" default:"53"`
    Domain     string        `env:"DNS_DOMAIN" default:"service.consul"`
    RefreshTTL time.Duration `env:"DNS_REFRESH_TTL" default:"30s"`
}
```

## Usage Examples

### Service Registration

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "github.com/user/core-backend/pkg/discovery"
)

func main() {
    disc, err := discovery.New(discovery.Config{
        Provider: "consul",
    })
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // Register service
    service := &discovery.Service{
        ID:      "user-service-1",
        Name:    "user-service",
        Version: "1.0.0",
        Address: "192.168.1.10",
        Port:    8080,
        Tags:    []string{"api", "v1"},
        Metadata: map[string]string{
            "region": "us-west",
        },
        Health: &discovery.HealthCheck{
            HTTP:     "http://localhost:8080/health",
            Interval: 10 * time.Second,
            Timeout:  2 * time.Second,
        },
    }

    if err := disc.Register(ctx, service); err != nil {
        log.Fatal(err)
    }

    // Graceful shutdown
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt)
    <-sigCh

    disc.Deregister(ctx, service.ID)
    disc.Close()
}
```

### Service Discovery

```go
func main() {
    disc, _ := discovery.New(cfg)

    ctx := context.Background()

    // Discover all instances
    instances, err := disc.Discover(ctx, "user-service")
    if err != nil {
        log.Fatal(err)
    }

    for _, inst := range instances {
        fmt.Printf("Found: %s at %s\n", inst.ID, inst.Endpoint())
    }

    // Resolve single instance (with load balancing)
    instance, err := disc.Resolve(ctx, "user-service")
    if err != nil {
        log.Fatal(err)
    }

    // Connect to instance
    conn, err := grpc.Dial(instance.Endpoint(), grpc.WithInsecure())
}
```

### Watch for Changes

```go
func main() {
    disc, _ := discovery.New(cfg)

    ctx := context.Background()

    // Watch for service changes
    disc.Watch(ctx, "user-service", func(instances []*discovery.Instance) {
        log.Printf("Service updated: %d instances\n", len(instances))

        // Update connection pool
        updateConnectionPool(instances)
    })
}
```

### Load Balancing

```go
import (
    "github.com/user/core-backend/pkg/discovery/balancer"
)

func main() {
    disc, _ := discovery.New(discovery.Config{
        Provider: "consul",
        Balancer: "weighted_round_robin",
    })

    // Or set balancer explicitly
    disc, _ := discovery.New(cfg,
        discovery.WithBalancer(balancer.NewWeightedRoundRobin()),
    )

    // Resolve uses the configured balancer
    instance, _ := disc.Resolve(ctx, "api-gateway")
}
```

### Kubernetes Discovery

```go
import (
    "github.com/user/core-backend/pkg/discovery/provider"
)

func main() {
    k8sProvider, err := provider.NewKubernetes(provider.KubernetesConfig{
        Namespace:     "production",
        LabelSelector: "app=user-service",
        InCluster:     true,
    })

    disc, _ := discovery.New(discovery.Config{},
        discovery.WithProvider(k8sProvider),
    )

    // Discovers pods matching selector
    instances, _ := disc.Discover(ctx, "user-service")
}
```

### With gRPC Client

```go
import (
    "github.com/user/core-backend/pkg/discovery"
    "google.golang.org/grpc"
)

func main() {
    disc, _ := discovery.New(cfg)

    // Create gRPC resolver
    resolver := discovery.NewGRPCResolver(disc)
    grpc.RegisterResolver(resolver)

    // Use discovery scheme
    conn, err := grpc.Dial("discovery:///user-service",
        grpc.WithInsecure(),
        grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
    )
}
```

## Dependencies

- **Required:** None (static provider uses stdlib)
- **Optional:**
  - `github.com/hashicorp/consul/api` for Consul
  - `go.etcd.io/etcd/client/v3` for etcd
  - `k8s.io/client-go` for Kubernetes

## Implementation Phases

### Phase 1: Core Interface & Static Provider
1. Define Discovery, Service interfaces
2. Static provider
3. Basic balancers

### Phase 2: Consul Provider
1. Service registration
2. Service discovery
3. Health check integration
4. Watch support

### Phase 3: Additional Providers
1. etcd provider
2. Kubernetes provider
3. DNS provider

### Phase 4: Load Balancing
1. Round-robin
2. Weighted round-robin
3. Least connections

### Phase 5: gRPC Integration
1. gRPC resolver
2. gRPC balancer

### Phase 6: Documentation
1. README
2. Examples

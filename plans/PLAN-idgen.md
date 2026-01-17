# Package Plan: pkg/idgen

## Overview

An ID generation package providing multiple algorithms for generating unique, distributed-safe identifiers. Supports UUID, ULID, Snowflake, and custom formats with features like sortability, time-encoding, and k-sortable IDs.

## Goals

1. **Multiple Algorithms** - UUID (v4, v7), ULID, Snowflake, NanoID, KSUID
2. **Distributed Safe** - No coordination required across nodes
3. **Sortable IDs** - Time-ordered for database indexing
4. **Customizable** - Custom alphabets, lengths, formats
5. **High Performance** - Lock-free where possible
6. **Type Safety** - Strongly typed ID types
7. **Zero Dependencies** - Pure Go implementation

## Architecture

```
pkg/idgen/
├── idgen.go              # Main generator interface
├── config.go             # Configuration
├── uuid.go               # UUID v4, v7
├── ulid.go               # ULID implementation
├── snowflake.go          # Snowflake IDs
├── nanoid.go             # NanoID
├── ksuid.go              # K-Sortable UID
├── shortid.go            # Short IDs
├── prefixed.go           # Prefixed IDs (usr_xxx)
├── types.go              # ID type definitions
├── encoding.go           # Encoding utilities
├── examples/
│   ├── basic/
│   ├── snowflake/
│   └── prefixed/
└── README.md
```

## Core Interfaces

```go
package idgen

import (
    "time"
)

// Generator generates unique IDs
type Generator interface {
    // Generate creates a new ID
    Generate() ID

    // GenerateN creates n IDs
    GenerateN(n int) []ID

    // Name returns the generator name
    Name() string
}

// ID represents a generated ID
type ID interface {
    // String returns string representation
    String() string

    // Bytes returns byte representation
    Bytes() []byte

    // Time returns embedded timestamp (if applicable)
    Time() (time.Time, bool)

    // Compare compares two IDs (-1, 0, 1)
    Compare(other ID) int
}

// Parser parses IDs from strings
type Parser interface {
    // Parse parses a string into an ID
    Parse(s string) (ID, error)

    // IsValid checks if string is a valid ID
    IsValid(s string) bool
}
```

## ID Types

### UUID

```go
// UUID represents a UUID
type UUID [16]byte

// V4 generates a random UUID v4
func V4() UUID

// V7 generates a time-ordered UUID v7
func V7() UUID

// Parse parses a UUID string
func ParseUUID(s string) (UUID, error)

func (u UUID) String() string        // "550e8400-e29b-41d4-a716-446655440000"
func (u UUID) Time() (time.Time, bool) // V7 only
func (u UUID) Version() int
```

### ULID

```go
// ULID is a Universally Unique Lexicographically Sortable Identifier
type ULID [16]byte

// NewULID generates a new ULID
func NewULID() ULID

// NewULIDWithTime generates a ULID with specific timestamp
func NewULIDWithTime(t time.Time) ULID

// Parse parses a ULID string
func ParseULID(s string) (ULID, error)

func (u ULID) String() string      // "01ARZ3NDEKTSV4RRFFQ69G5FAV"
func (u ULID) Time() time.Time     // Embedded timestamp
func (u ULID) Entropy() []byte     // Random component
```

### Snowflake

```go
// Snowflake is a Twitter Snowflake ID
type Snowflake int64

// SnowflakeGenerator generates Snowflake IDs
type SnowflakeGenerator struct {
    nodeID    int64
    epoch     time.Time
    sequence  int64
    lastTime  int64
}

// NewSnowflake creates a Snowflake generator
func NewSnowflake(nodeID int64, opts ...SnowflakeOption) *SnowflakeGenerator

// Options
func WithEpoch(epoch time.Time) SnowflakeOption // Default: Twitter epoch
func WithNodeBits(bits int) SnowflakeOption     // Default: 10
func WithSequenceBits(bits int) SnowflakeOption // Default: 12

func (g *SnowflakeGenerator) Generate() Snowflake
func (s Snowflake) String() string      // "1382971839183872"
func (s Snowflake) Time() time.Time     // Embedded timestamp
func (s Snowflake) NodeID() int64       // Node that generated it
func (s Snowflake) Sequence() int64     // Sequence number
```

### NanoID

```go
// NanoID is a tiny, URL-safe, unique string ID
type NanoID string

// NewNanoID generates a NanoID with default settings
func NewNanoID() NanoID

// NewNanoIDWithLength generates with specific length
func NewNanoIDWithLength(length int) NanoID

// NewNanoIDWithAlphabet generates with custom alphabet
func NewNanoIDWithAlphabet(length int, alphabet string) NanoID

func (n NanoID) String() string // "V1StGXR8_Z5jdHi6B-myT"

// Default: 21 chars, URL-safe alphabet
// Collision probability: 1 billion IDs needed for 1% collision
```

### KSUID

```go
// KSUID is a K-Sortable Unique ID
type KSUID [20]byte

// NewKSUID generates a new KSUID
func NewKSUID() KSUID

func (k KSUID) String() string    // "0ujsswThIGTUYm2K8FjOOfXtY1K"
func (k KSUID) Time() time.Time   // Timestamp (4 bytes)
func (k KSUID) Payload() []byte   // Random payload (16 bytes)
```

### Prefixed IDs

```go
// PrefixedID combines a prefix with an underlying ID
type PrefixedID struct {
    Prefix string
    ID     ID
}

// PrefixedGenerator wraps a generator with a prefix
type PrefixedGenerator struct {
    prefix    string
    separator string
    generator Generator
}

func NewPrefixedGenerator(prefix string, gen Generator, opts ...PrefixedOption) *PrefixedGenerator

// Options
func WithSeparator(sep string) PrefixedOption // Default: "_"

// Example output: "usr_01ARZ3NDEKTSV4RRFFQ69G5FAV"
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/user/core-backend/pkg/idgen"
)

func main() {
    // UUID v4 (random)
    uuid := idgen.V4()
    fmt.Println(uuid.String()) // "550e8400-e29b-41d4-a716-446655440000"

    // UUID v7 (time-ordered)
    uuidv7 := idgen.V7()
    fmt.Println(uuidv7.String()) // "018e5e3c-7b1a-7d4e-8f1a-2b3c4d5e6f70"
    fmt.Println(uuidv7.Time())   // 2024-01-15 10:30:00

    // ULID (sortable)
    ulid := idgen.NewULID()
    fmt.Println(ulid.String()) // "01ARZ3NDEKTSV4RRFFQ69G5FAV"
    fmt.Println(ulid.Time())   // Embedded timestamp

    // NanoID (short, URL-safe)
    nanoid := idgen.NewNanoID()
    fmt.Println(nanoid.String()) // "V1StGXR8_Z5jdHi6B-myT"
}
```

### Snowflake IDs

```go
func main() {
    // Create generator with node ID
    gen := idgen.NewSnowflake(1, // Node ID 1
        idgen.WithEpoch(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
    )

    // Generate IDs
    id := gen.Generate()
    fmt.Println(id.String())    // "1382971839183872"
    fmt.Println(id.Time())      // Embedded timestamp
    fmt.Println(id.NodeID())    // 1
    fmt.Println(id.Sequence())  // Sequence within millisecond
}
```

### Prefixed IDs

```go
func main() {
    // Create prefixed generators for different entities
    userIDGen := idgen.NewPrefixedGenerator("usr", idgen.NewULIDGenerator())
    orderIDGen := idgen.NewPrefixedGenerator("ord", idgen.NewULIDGenerator())
    txIDGen := idgen.NewPrefixedGenerator("txn", idgen.NewULIDGenerator())

    userID := userIDGen.Generate()
    fmt.Println(userID.String()) // "usr_01ARZ3NDEKTSV4RRFFQ69G5FAV"

    orderID := orderIDGen.Generate()
    fmt.Println(orderID.String()) // "ord_01ARZ3NDEKTSV4RRFFQ69G5FAV"
}
```

### Custom NanoID

```go
func main() {
    // Custom length
    id := idgen.NewNanoIDWithLength(10)
    fmt.Println(id) // "V1StGXR8_Z"

    // Custom alphabet (numeric only)
    numericID := idgen.NewNanoIDWithAlphabet(12, "0123456789")
    fmt.Println(numericID) // "839271649102"

    // Custom alphabet (lowercase + numbers)
    customID := idgen.NewNanoIDWithAlphabet(8, "abcdefghijklmnopqrstuvwxyz0123456789")
    fmt.Println(customID) // "k7x2m9p4"
}
```

### Parsing IDs

```go
func main() {
    // Parse UUID
    uuid, err := idgen.ParseUUID("550e8400-e29b-41d4-a716-446655440000")
    if err != nil {
        log.Fatal(err)
    }

    // Parse ULID
    ulid, err := idgen.ParseULID("01ARZ3NDEKTSV4RRFFQ69G5FAV")
    if err != nil {
        log.Fatal(err)
    }

    // Validate
    if idgen.IsValidUUID(input) {
        // ...
    }
}
```

### Thread-Safe Generator

```go
func main() {
    gen := idgen.NewULIDGenerator()

    // Safe for concurrent use
    var wg sync.WaitGroup
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            id := gen.Generate()
            // Use id...
        }()
    }
    wg.Wait()
}
```

### Database Integration

```go
import (
    "database/sql/driver"
)

// UUID implements sql.Scanner and driver.Valuer
func (u *UUID) Scan(src interface{}) error
func (u UUID) Value() (driver.Value, error)

// Usage in models
type User struct {
    ID   idgen.UUID `db:"id"`
    Name string     `db:"name"`
}

// Insert
user := User{
    ID:   idgen.V4(),
    Name: "John",
}
db.Exec("INSERT INTO users (id, name) VALUES ($1, $2)", user.ID, user.Name)
```

## ID Comparison

| Type | Length | Sortable | Time-Based | Collision Resistance |
|------|--------|----------|------------|---------------------|
| UUID v4 | 36 chars | No | No | 2^122 |
| UUID v7 | 36 chars | Yes | Yes | 2^62 per ms |
| ULID | 26 chars | Yes | Yes | 2^80 per ms |
| Snowflake | 18 digits | Yes | Yes | 4096/ms/node |
| NanoID | 21 chars | No | No | Configurable |
| KSUID | 27 chars | Yes | Yes | 2^128 per sec |

## Dependencies

- **Required:** None (pure Go)

## Implementation Phases

### Phase 1: Core Types
1. UUID v4 implementation
2. Basic parsing and validation
3. String/byte conversions

### Phase 2: Time-Based IDs
1. UUID v7
2. ULID
3. KSUID

### Phase 3: Distributed IDs
1. Snowflake implementation
2. Node ID management
3. Clock rollback handling

### Phase 4: Custom IDs
1. NanoID
2. Prefixed IDs
3. Custom alphabets

### Phase 5: Database Integration
1. SQL Scanner/Valuer
2. JSON marshaling
3. BSON support

### Phase 6: Documentation
1. README
2. Comparison guide
3. Best practices

# Package Plan: pkg/search

## Overview

A unified search client package supporting multiple search engines (Elasticsearch, OpenSearch, Meilisearch, Typesense). Provides a consistent API for indexing, searching, and managing documents with support for facets, filters, and aggregations.

## Goals

1. **Multiple Backends** - Elasticsearch, OpenSearch, Meilisearch, Typesense
2. **Unified Interface** - Single API for all search engines
3. **Type-Safe Queries** - Builder pattern for query construction
4. **Faceted Search** - Support facets and aggregations
5. **Bulk Operations** - Efficient bulk indexing
6. **Highlighting** - Search result highlighting
7. **Suggestions** - Autocomplete and suggestions

## Architecture

```
pkg/search/
├── search.go             # Core Client interface
├── config.go             # Configuration
├── options.go            # Functional options
├── query.go              # Query builder
├── result.go             # Search results
├── index.go              # Index management
├── errors.go             # Custom error types
├── provider/
│   ├── provider.go       # Provider interface
│   ├── elasticsearch.go  # Elasticsearch/OpenSearch
│   ├── meilisearch.go    # Meilisearch
│   └── typesense.go      # Typesense
├── examples/
│   ├── basic/
│   ├── elasticsearch/
│   ├── meilisearch/
│   └── faceted-search/
└── README.md
```

## Core Interfaces

```go
package search

import (
    "context"
)

// Client provides search operations
type Client interface {
    // Index indexes a document
    Index(ctx context.Context, index string, id string, doc interface{}) error

    // BulkIndex indexes multiple documents
    BulkIndex(ctx context.Context, index string, docs []Document) error

    // Get retrieves a document by ID
    Get(ctx context.Context, index, id string, dst interface{}) error

    // Delete removes a document
    Delete(ctx context.Context, index, id string) error

    // Search executes a search query
    Search(ctx context.Context, index string, query *Query) (*Result, error)

    // Suggest returns autocomplete suggestions
    Suggest(ctx context.Context, index, field, prefix string, opts ...SuggestOption) ([]string, error)

    // CreateIndex creates a new index
    CreateIndex(ctx context.Context, index string, mapping *Mapping) error

    // DeleteIndex deletes an index
    DeleteIndex(ctx context.Context, index string) error

    // IndexExists checks if index exists
    IndexExists(ctx context.Context, index string) (bool, error)

    // Close releases resources
    Close() error
}

// Document for bulk indexing
type Document struct {
    ID   string
    Data interface{}
}

// Query represents a search query
type Query struct {
    Text       string
    Fields     []string
    Filters    []Filter
    Facets     []string
    Sort       []Sort
    From       int
    Size       int
    Highlight  *Highlight
}

// Filter represents a search filter
type Filter struct {
    Field    string
    Operator string // "eq", "ne", "gt", "gte", "lt", "lte", "in", "range"
    Value    interface{}
}

// Sort represents sort order
type Sort struct {
    Field string
    Order string // "asc" or "desc"
}

// Highlight configuration
type Highlight struct {
    Fields    []string
    PreTag    string
    PostTag   string
}

// Result represents search results
type Result struct {
    Hits       []Hit
    Total      int64
    Facets     map[string][]FacetValue
    Took       time.Duration
}

// Hit represents a single search hit
type Hit struct {
    ID         string
    Score      float64
    Source     json.RawMessage
    Highlights map[string][]string
}

// Scan scans hit source into struct
func (h *Hit) Scan(dst interface{}) error

// FacetValue represents a facet bucket
type FacetValue struct {
    Value string
    Count int64
}

// Mapping defines index schema
type Mapping struct {
    Properties map[string]FieldMapping
    Settings   map[string]interface{}
}

// FieldMapping defines field configuration
type FieldMapping struct {
    Type       string // "text", "keyword", "integer", "float", "boolean", "date", "geo_point"
    Analyzer   string
    Searchable bool
    Sortable   bool
    Filterable bool
}
```

## Query Builder

```go
// NewQuery creates a query builder
func NewQuery() *QueryBuilder

type QueryBuilder struct {
    query *Query
}

func (b *QueryBuilder) Text(text string) *QueryBuilder
func (b *QueryBuilder) Fields(fields ...string) *QueryBuilder
func (b *QueryBuilder) Filter(field, op string, value interface{}) *QueryBuilder
func (b *QueryBuilder) FilterEq(field string, value interface{}) *QueryBuilder
func (b *QueryBuilder) FilterIn(field string, values ...interface{}) *QueryBuilder
func (b *QueryBuilder) FilterRange(field string, min, max interface{}) *QueryBuilder
func (b *QueryBuilder) Facet(fields ...string) *QueryBuilder
func (b *QueryBuilder) Sort(field, order string) *QueryBuilder
func (b *QueryBuilder) SortAsc(field string) *QueryBuilder
func (b *QueryBuilder) SortDesc(field string) *QueryBuilder
func (b *QueryBuilder) From(from int) *QueryBuilder
func (b *QueryBuilder) Size(size int) *QueryBuilder
func (b *QueryBuilder) Highlight(fields ...string) *QueryBuilder
func (b *QueryBuilder) Build() *Query
```

## Configuration

```go
// Config holds search client configuration
type Config struct {
    // Provider: "elasticsearch", "opensearch", "meilisearch", "typesense"
    Provider string `env:"SEARCH_PROVIDER" default:"elasticsearch"`
}

// ElasticsearchConfig for Elasticsearch/OpenSearch
type ElasticsearchConfig struct {
    URLs     []string `env:"ELASTICSEARCH_URLS" default:"http://localhost:9200"`
    Username string   `env:"ELASTICSEARCH_USERNAME"`
    Password string   `env:"ELASTICSEARCH_PASSWORD"`
    APIKey   string   `env:"ELASTICSEARCH_API_KEY"`
    CloudID  string   `env:"ELASTICSEARCH_CLOUD_ID"`
}

// MeilisearchConfig for Meilisearch
type MeilisearchConfig struct {
    URL    string `env:"MEILISEARCH_URL" default:"http://localhost:7700"`
    APIKey string `env:"MEILISEARCH_API_KEY"`
}

// TypesenseConfig for Typesense
type TypesenseConfig struct {
    URLs   []string `env:"TYPESENSE_URLS" default:"http://localhost:8108"`
    APIKey string   `env:"TYPESENSE_API_KEY" required:"true"`
}
```

## Usage Examples

### Basic Search

```go
package main

import (
    "context"
    "github.com/user/core-backend/pkg/search"
)

type Product struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Category    string   `json:"category"`
    Price       float64  `json:"price"`
    Tags        []string `json:"tags"`
}

func main() {
    client, _ := search.New(search.Config{
        Provider: "elasticsearch",
    })
    defer client.Close()

    ctx := context.Background()

    // Index a document
    product := Product{
        ID:          "prod-1",
        Name:        "Wireless Mouse",
        Description: "Ergonomic wireless mouse with USB receiver",
        Category:    "Electronics",
        Price:       29.99,
        Tags:        []string{"wireless", "mouse", "computer"},
    }

    client.Index(ctx, "products", product.ID, product)

    // Search
    query := search.NewQuery().
        Text("wireless mouse").
        Fields("name", "description").
        FilterEq("category", "Electronics").
        FilterRange("price", 0, 50).
        SortDesc("_score").
        Size(10).
        Build()

    result, _ := client.Search(ctx, "products", query)

    for _, hit := range result.Hits {
        var p Product
        hit.Scan(&p)
        fmt.Printf("%s: %s ($%.2f)\n", p.ID, p.Name, p.Price)
    }
}
```

### Faceted Search

```go
query := search.NewQuery().
    Text("laptop").
    Facet("category", "brand", "price_range").
    Build()

result, _ := client.Search(ctx, "products", query)

// Display facets
for facetName, values := range result.Facets {
    fmt.Printf("%s:\n", facetName)
    for _, v := range values {
        fmt.Printf("  %s (%d)\n", v.Value, v.Count)
    }
}
// Output:
// category:
//   Electronics (45)
//   Computers (32)
// brand:
//   Apple (20)
//   Dell (15)
```

### Autocomplete

```go
suggestions, _ := client.Suggest(ctx, "products", "name", "wire",
    search.WithSuggestLimit(5),
)
// ["Wireless Mouse", "Wireless Keyboard", "Wireless Charger"]
```

## Dependencies

- **Required:** None (interface only)
- **Optional:**
  - `github.com/elastic/go-elasticsearch/v8` for Elasticsearch
  - `github.com/meilisearch/meilisearch-go` for Meilisearch
  - `github.com/typesense/typesense-go` for Typesense

## Implementation Phases

### Phase 1: Core Interface & Elasticsearch
1. Define Client interface
2. Query builder
3. Elasticsearch provider

### Phase 2: Additional Providers
1. Meilisearch provider
2. Typesense provider

### Phase 3: Advanced Features
1. Faceted search
2. Highlighting
3. Suggestions/autocomplete

### Phase 4: Documentation
1. README
2. Examples for each provider

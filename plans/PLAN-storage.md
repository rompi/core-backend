# Package Plan: pkg/storage

## Overview

A unified file storage abstraction supporting multiple backends (local filesystem, S3, Google Cloud Storage, Azure Blob Storage). Provides a consistent API for uploading, downloading, and managing files with support for presigned URLs, metadata, and streaming.

## Goals

1. **Unified Interface** - Single API for all storage backends
2. **Multiple Backends** - Local, S3, GCS, Azure Blob
3. **Streaming Support** - Handle large files without memory issues
4. **Presigned URLs** - Generate temporary access URLs
5. **Metadata** - Store and retrieve file metadata
6. **Directory Operations** - List, copy, move, delete directories
7. **Zero Core Dependencies** - Backend SDKs are optional

## Architecture

```
pkg/storage/
├── storage.go           # Core Storage interface
├── config.go            # Configuration with env support
├── options.go           # Functional options
├── file.go              # File type with metadata
├── errors.go            # Custom error types
├── local/
│   ├── local.go         # Local filesystem implementation
│   ├── config.go
│   └── local_test.go
├── s3/
│   ├── s3.go            # AWS S3 implementation
│   ├── config.go
│   └── s3_test.go
├── gcs/
│   ├── gcs.go           # Google Cloud Storage implementation
│   ├── config.go
│   └── gcs_test.go
├── azure/
│   ├── azure.go         # Azure Blob Storage implementation
│   ├── config.go
│   └── azure_test.go
├── middleware/
│   ├── logging.go       # Logging middleware
│   ├── metrics.go       # Metrics middleware
│   └── validation.go    # File validation middleware
├── examples/
│   ├── basic/
│   ├── s3/
│   ├── presigned-urls/
│   └── streaming/
└── README.md
```

## Core Interfaces

```go
package storage

import (
    "context"
    "io"
    "time"
)

// Storage defines the unified storage interface
type Storage interface {
    // Upload stores a file
    Upload(ctx context.Context, path string, reader io.Reader, opts ...UploadOption) (*File, error)

    // Download retrieves a file
    Download(ctx context.Context, path string) (io.ReadCloser, error)

    // Delete removes a file
    Delete(ctx context.Context, path string) error

    // Exists checks if a file exists
    Exists(ctx context.Context, path string) (bool, error)

    // Stat returns file metadata without downloading
    Stat(ctx context.Context, path string) (*File, error)

    // List returns files in a directory
    List(ctx context.Context, prefix string, opts ...ListOption) (FileIterator, error)

    // Copy copies a file to a new location
    Copy(ctx context.Context, src, dst string) error

    // Move moves a file to a new location
    Move(ctx context.Context, src, dst string) error

    // Close releases resources
    Close() error
}

// URLGenerator generates presigned URLs
type URLGenerator interface {
    // PresignedUploadURL generates a URL for direct upload
    PresignedUploadURL(ctx context.Context, path string, expiry time.Duration) (string, error)

    // PresignedDownloadURL generates a URL for direct download
    PresignedDownloadURL(ctx context.Context, path string, expiry time.Duration) (string, error)
}

// File represents a stored file with metadata
type File struct {
    // Path is the file location
    Path string

    // Size in bytes
    Size int64

    // ContentType (MIME type)
    ContentType string

    // LastModified timestamp
    LastModified time.Time

    // ETag for caching/versioning
    ETag string

    // Metadata holds custom key-value pairs
    Metadata map[string]string
}

// FileIterator iterates over files
type FileIterator interface {
    // Next returns the next file or io.EOF
    Next() (*File, error)

    // Close releases iterator resources
    Close() error
}

// UploadOption configures upload behavior
type UploadOption func(*uploadOptions)

// ListOption configures list behavior
type ListOption func(*listOptions)
```

## Configuration

```go
// Config holds storage configuration
type Config struct {
    // Backend type: "local", "s3", "gcs", "azure"
    Backend string `env:"STORAGE_BACKEND" default:"local"`

    // Base path/bucket prefix
    BasePath string `env:"STORAGE_BASE_PATH" default:""`

    // Public base URL for serving files
    PublicURL string `env:"STORAGE_PUBLIC_URL" default:""`
}

// UploadOptions for upload configuration
type uploadOptions struct {
    ContentType     string
    Metadata        map[string]string
    CacheControl    string
    ContentEncoding string
    ACL             string
}

// ListOptions for list configuration
type listOptions struct {
    Recursive bool
    MaxResults int
    StartAfter string
}
```

## Backend Configurations

### Local Filesystem

```go
type LocalConfig struct {
    // Root directory for storage
    RootDir string `env:"STORAGE_LOCAL_ROOT" default:"./uploads"`

    // Directory permissions
    DirMode os.FileMode `env:"STORAGE_LOCAL_DIR_MODE" default:"0755"`

    // File permissions
    FileMode os.FileMode `env:"STORAGE_LOCAL_FILE_MODE" default:"0644"`

    // Create directories automatically
    AutoCreate bool `env:"STORAGE_LOCAL_AUTO_CREATE" default:"true"`
}
```

### AWS S3

```go
type S3Config struct {
    // Bucket name
    Bucket string `env:"STORAGE_S3_BUCKET" required:"true"`

    // Region
    Region string `env:"STORAGE_S3_REGION" default:"us-east-1"`

    // Endpoint (for S3-compatible services like MinIO)
    Endpoint string `env:"STORAGE_S3_ENDPOINT" default:""`

    // Use path-style addressing (for MinIO)
    ForcePathStyle bool `env:"STORAGE_S3_FORCE_PATH_STYLE" default:"false"`

    // Credentials (if not using IAM/environment)
    AccessKeyID     string `env:"STORAGE_S3_ACCESS_KEY_ID" default:""`
    SecretAccessKey string `env:"STORAGE_S3_SECRET_ACCESS_KEY" default:""`

    // Default ACL for uploads
    DefaultACL string `env:"STORAGE_S3_DEFAULT_ACL" default:"private"`

    // Server-side encryption
    ServerSideEncryption string `env:"STORAGE_S3_SSE" default:""`

    // Upload part size for multipart uploads
    PartSize int64 `env:"STORAGE_S3_PART_SIZE" default:"5242880"` // 5MB

    // Concurrent upload workers
    Concurrency int `env:"STORAGE_S3_CONCURRENCY" default:"5"`
}
```

### Google Cloud Storage

```go
type GCSConfig struct {
    // Bucket name
    Bucket string `env:"STORAGE_GCS_BUCKET" required:"true"`

    // Project ID
    ProjectID string `env:"STORAGE_GCS_PROJECT_ID" default:""`

    // Credentials JSON file path
    CredentialsFile string `env:"STORAGE_GCS_CREDENTIALS_FILE" default:""`

    // Credentials JSON string
    CredentialsJSON string `env:"STORAGE_GCS_CREDENTIALS_JSON" default:""`

    // Default ACL
    DefaultACL string `env:"STORAGE_GCS_DEFAULT_ACL" default:"private"`

    // Chunk size for resumable uploads
    ChunkSize int `env:"STORAGE_GCS_CHUNK_SIZE" default:"8388608"` // 8MB
}
```

### Azure Blob Storage

```go
type AzureConfig struct {
    // Container name
    Container string `env:"STORAGE_AZURE_CONTAINER" required:"true"`

    // Account name
    AccountName string `env:"STORAGE_AZURE_ACCOUNT_NAME" required:"true"`

    // Account key
    AccountKey string `env:"STORAGE_AZURE_ACCOUNT_KEY" default:""`

    // Connection string (alternative to account name/key)
    ConnectionString string `env:"STORAGE_AZURE_CONNECTION_STRING" default:""`

    // SAS token (alternative authentication)
    SASToken string `env:"STORAGE_AZURE_SAS_TOKEN" default:""`

    // Default access tier
    AccessTier string `env:"STORAGE_AZURE_ACCESS_TIER" default:"Hot"`
}
```

## Local Implementation

```go
// Local implements Storage using the local filesystem
type Local struct {
    rootDir  string
    dirMode  os.FileMode
    fileMode os.FileMode
}

// NewLocal creates a local storage
func NewLocal(cfg LocalConfig, opts ...Option) (*Local, error)

// Upload stores a file locally
func (l *Local) Upload(ctx context.Context, path string, reader io.Reader, opts ...UploadOption) (*File, error)

// Download retrieves a file
func (l *Local) Download(ctx context.Context, path string) (io.ReadCloser, error)
```

## Upload Options

```go
// WithContentType sets the MIME type
func WithContentType(ct string) UploadOption

// WithMetadata sets custom metadata
func WithMetadata(meta map[string]string) UploadOption

// WithCacheControl sets cache headers
func WithCacheControl(cc string) UploadOption

// WithACL sets access control
func WithACL(acl string) UploadOption

// WithContentEncoding sets encoding (e.g., "gzip")
func WithContentEncoding(enc string) UploadOption
```

## List Options

```go
// WithRecursive enables recursive listing
func WithRecursive(recursive bool) ListOption

// WithMaxResults limits results per page
func WithMaxResults(max int) ListOption

// WithStartAfter sets cursor for pagination
func WithStartAfter(cursor string) ListOption
```

## Error Handling

```go
var (
    // ErrNotFound is returned when file doesn't exist
    ErrNotFound = errors.New("storage: file not found")

    // ErrAlreadyExists is returned when file already exists
    ErrAlreadyExists = errors.New("storage: file already exists")

    // ErrAccessDenied is returned on permission errors
    ErrAccessDenied = errors.New("storage: access denied")

    // ErrInvalidPath is returned for invalid file paths
    ErrInvalidPath = errors.New("storage: invalid path")

    // ErrTooLarge is returned when file exceeds size limit
    ErrTooLarge = errors.New("storage: file too large")
)

// IsNotFound checks if error is ErrNotFound
func IsNotFound(err error) bool

// IsAccessDenied checks if error is ErrAccessDenied
func IsAccessDenied(err error) bool
```

## Middleware

```go
// Middleware wraps a Storage
type Middleware func(Storage) Storage

// Logging adds logging to operations
func Logging(logger Logger) Middleware

// Metrics adds metrics collection
func Metrics(collector MetricsCollector) Middleware

// Validation validates files before upload
func Validation(rules ValidationRules) Middleware

// ValidationRules for file validation
type ValidationRules struct {
    // Maximum file size in bytes
    MaxSize int64

    // Allowed MIME types
    AllowedTypes []string

    // Blocked file extensions
    BlockedExtensions []string
}
```

## Usage Examples

### Basic Upload/Download

```go
package main

import (
    "context"
    "os"
    "github.com/user/core-backend/pkg/storage"
    "github.com/user/core-backend/pkg/storage/local"
)

func main() {
    // Create local storage
    store, err := local.New(local.Config{
        RootDir: "./uploads",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer store.Close()

    ctx := context.Background()

    // Upload a file
    file, _ := os.Open("document.pdf")
    defer file.Close()

    uploaded, err := store.Upload(ctx, "documents/report.pdf", file,
        storage.WithContentType("application/pdf"),
        storage.WithMetadata(map[string]string{
            "author": "John Doe",
        }),
    )

    // Download a file
    reader, err := store.Download(ctx, "documents/report.pdf")
    if err != nil {
        if storage.IsNotFound(err) {
            // Handle not found
        }
    }
    defer reader.Close()

    // Copy to output
    io.Copy(os.Stdout, reader)
}
```

### S3 Storage

```go
import (
    "github.com/user/core-backend/pkg/storage/s3"
)

func main() {
    store, err := s3.New(s3.Config{
        Bucket:   "my-bucket",
        Region:   "us-west-2",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer store.Close()

    // Same interface as local storage
    uploaded, err := store.Upload(ctx, "images/photo.jpg", reader,
        storage.WithContentType("image/jpeg"),
        storage.WithACL("public-read"),
    )

    // Generate presigned URL
    url, err := store.PresignedDownloadURL(ctx, "images/photo.jpg", time.Hour)
    fmt.Println("Download URL:", url)
}
```

### Presigned URLs for Direct Upload

```go
func main() {
    store, _ := s3.New(cfg)

    // Generate presigned upload URL for client
    uploadURL, err := store.PresignedUploadURL(ctx, "uploads/user-file.pdf", 15*time.Minute)

    // Client can now upload directly to S3
    // POST to uploadURL with file data
}
```

### Listing Files

```go
func main() {
    store, _ := local.New(cfg)

    // List files in directory
    iter, err := store.List(ctx, "documents/",
        storage.WithRecursive(true),
        storage.WithMaxResults(100),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer iter.Close()

    for {
        file, err := iter.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("%s (%d bytes)\n", file.Path, file.Size)
    }
}
```

### With Middleware

```go
func main() {
    baseStore, _ := s3.New(cfg)

    // Wrap with middleware
    store := storage.Chain(
        storage.Logging(logger),
        storage.Metrics(prometheus),
        storage.Validation(storage.ValidationRules{
            MaxSize:           10 * 1024 * 1024, // 10MB
            AllowedTypes:      []string{"image/jpeg", "image/png", "application/pdf"},
            BlockedExtensions: []string{".exe", ".sh", ".bat"},
        }),
    )(baseStore)

    // Use wrapped store
    _, err := store.Upload(ctx, "file.exe", reader) // ErrValidation
}
```

### MinIO (S3-Compatible)

```go
func main() {
    store, err := s3.New(s3.Config{
        Bucket:         "my-bucket",
        Endpoint:       "http://localhost:9000",
        ForcePathStyle: true,
        AccessKeyID:    "minioadmin",
        SecretAccessKey: "minioadmin",
    })
    // Use same interface
}
```

## Health Check

```go
// HealthCheck returns a health check function
func (s *S3) HealthCheck() func(ctx context.Context) error {
    return func(ctx context.Context) error {
        _, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
            Bucket: aws.String(s.bucket),
        })
        return err
    }
}
```

## Observability Hooks

```go
// Hook interface for observability
type Hook interface {
    BeforeUpload(ctx context.Context, path string, size int64)
    AfterUpload(ctx context.Context, path string, file *File, err error)
    BeforeDownload(ctx context.Context, path string)
    AfterDownload(ctx context.Context, path string, size int64, err error)
}

// WithHook adds observability hooks
func WithHook(hook Hook) Option
```

## Dependencies

- **Required:** None (local implementation)
- **Optional:**
  - `github.com/aws/aws-sdk-go-v2` for S3
  - `cloud.google.com/go/storage` for GCS
  - `github.com/Azure/azure-sdk-for-go` for Azure

## Test Coverage Requirements

- Unit tests for all public functions
- Integration tests with localstack/MinIO for S3
- Integration tests with emulators for GCS/Azure
- Benchmark tests for large files
- 80%+ coverage target

## Implementation Phases

### Phase 1: Core Interface & Local Implementation
1. Define Storage interface
2. Implement local filesystem storage
3. Add file metadata support
4. Write comprehensive tests

### Phase 2: S3 Implementation
1. Implement S3 storage
2. Add multipart upload support
3. Presigned URL generation
4. Integration tests with localstack

### Phase 3: GCS Implementation
1. Implement GCS storage
2. Resumable upload support
3. Integration tests

### Phase 4: Azure Implementation
1. Implement Azure Blob storage
2. Access tier support
3. Integration tests

### Phase 5: Advanced Features
1. Middleware system
2. File validation
3. Observability hooks
4. Streaming utilities

### Phase 6: Documentation & Examples
1. README with full documentation
2. Example for each backend
3. Presigned URL example
4. Middleware example

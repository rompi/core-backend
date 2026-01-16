package gateway

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

// CompressionConfig configures the compression middleware.
type CompressionConfig struct {
	// Level is the compression level (1-9, default: gzip.DefaultCompression).
	Level int

	// MinSize is the minimum response size to compress (default: 1024 bytes).
	MinSize int

	// ContentTypes is a list of content types to compress.
	// If empty, all text-based content types are compressed.
	ContentTypes []string
}

// DefaultCompressionConfig returns default compression configuration.
func DefaultCompressionConfig() CompressionConfig {
	return CompressionConfig{
		Level:   gzip.DefaultCompression,
		MinSize: 1024,
		ContentTypes: []string{
			"application/json",
			"application/xml",
			"text/html",
			"text/plain",
			"text/css",
			"text/javascript",
			"application/javascript",
		},
	}
}

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		return w
	},
}

// CompressionMiddleware creates response compression middleware.
func CompressionMiddleware() Middleware {
	return CompressionMiddlewareWithConfig(DefaultCompressionConfig())
}

// CompressionMiddlewareWithConfig creates compression middleware with config.
func CompressionMiddlewareWithConfig(config CompressionConfig) Middleware {
	contentTypes := make(map[string]bool)
	for _, ct := range config.ContentTypes {
		contentTypes[ct] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if client accepts gzip
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			// Create gzip response writer
			gz := gzipWriterPool.Get().(*gzip.Writer)
			gz.Reset(w)
			defer func() {
				gz.Close()
				gzipWriterPool.Put(gz)
			}()

			gzw := &gzipResponseWriter{
				ResponseWriter: w,
				writer:         gz,
				minSize:        config.MinSize,
				contentTypes:   contentTypes,
			}

			// Set headers
			w.Header().Set("Vary", "Accept-Encoding")

			next.ServeHTTP(gzw, r)

			// Flush if compression was used
			if gzw.compressed {
				gz.Flush()
			}
		})
	}
}

// gzipResponseWriter wraps http.ResponseWriter with gzip compression.
type gzipResponseWriter struct {
	http.ResponseWriter
	writer       *gzip.Writer
	minSize      int
	contentTypes map[string]bool
	buf          []byte
	compressed   bool
	wroteHeader  bool
}

// WriteHeader captures the status code and sets compression headers if needed.
func (w *gzipResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	// Check if content type should be compressed
	contentType := w.Header().Get("Content-Type")
	if contentType != "" {
		// Extract base content type (without charset, etc.)
		if idx := strings.Index(contentType, ";"); idx > 0 {
			contentType = strings.TrimSpace(contentType[:idx])
		}

		if len(w.contentTypes) > 0 && !w.contentTypes[contentType] {
			w.ResponseWriter.WriteHeader(code)
			return
		}
	}

	w.ResponseWriter.WriteHeader(code)
}

// Write compresses data if appropriate.
func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	// Buffer small responses
	if !w.compressed && len(w.buf)+len(b) < w.minSize {
		w.buf = append(w.buf, b...)
		return len(b), nil
	}

	// Start compression if we have buffered data
	if !w.compressed && len(w.buf) > 0 {
		w.startCompression()
		if _, err := w.writer.Write(w.buf); err != nil {
			return 0, err
		}
		w.buf = nil
	}

	if !w.compressed {
		w.startCompression()
	}

	return w.writer.Write(b)
}

// startCompression sets compression headers and marks as compressed.
func (w *gzipResponseWriter) startCompression() {
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Del("Content-Length") // Length changes with compression
	w.compressed = true
}

// Flush writes any buffered data and flushes the gzip writer.
func (w *gzipResponseWriter) Flush() {
	// Write any buffered data without compression if too small
	if !w.compressed && len(w.buf) > 0 {
		w.ResponseWriter.Write(w.buf)
		w.buf = nil
	}

	if w.compressed {
		w.writer.Flush()
	}

	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap returns the original ResponseWriter for http.ResponseController.
func (w *gzipResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

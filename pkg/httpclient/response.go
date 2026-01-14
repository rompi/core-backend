package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Response wraps http.Response with convenient helper methods
// for reading and decoding response bodies.
type Response struct {
	*http.Response
	body []byte // cached body for multiple reads
}

// JSON decodes the response body as JSON into the provided value.
// The response body is cached, so this method can be called multiple times.
//
// Returns an error if the body cannot be read or decoded.
func (r *Response) JSON(v interface{}) error {
	if err := r.cacheBody(); err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if err := json.Unmarshal(r.body, v); err != nil {
		return fmt.Errorf("decoding JSON: %w", err)
	}

	return nil
}

// String returns the response body as a string.
// The response body is cached, so this method can be called multiple times.
//
// Returns an error if the body cannot be read.
func (r *Response) String() (string, error) {
	bytes, err := r.Bytes()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Bytes returns the response body as a byte slice.
// The response body is cached, so this method can be called multiple times.
//
// Returns an error if the body cannot be read.
func (r *Response) Bytes() ([]byte, error) {
	if err := r.cacheBody(); err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	return r.body, nil
}

// IsSuccess returns true if the response status code is in the 2xx range.
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsClientError returns true if the response status code is in the 4xx range.
func (r *Response) IsClientError() bool {
	return r.StatusCode >= 400 && r.StatusCode < 500
}

// IsServerError returns true if the response status code is in the 5xx range.
func (r *Response) IsServerError() bool {
	return r.StatusCode >= 500 && r.StatusCode < 600
}

// cacheBody reads and caches the response body if not already cached.
// This allows the body to be read multiple times without consuming the stream.
func (r *Response) cacheBody() error {
	if r.body != nil {
		return nil // already cached
	}

	if r.Body == nil {
		return nil // no body to cache
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	r.body = body
	return nil
}

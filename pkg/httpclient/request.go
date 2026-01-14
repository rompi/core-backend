package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// RequestBuilder provides a fluent API for building and executing HTTP requests.
// It is not safe for concurrent use and should not be reused after calling Do().
type RequestBuilder struct {
	client  *Client
	ctx     context.Context
	method  string
	url     string
	headers http.Header
	query   url.Values
	body    io.Reader
}

// newRequestBuilder creates a new request builder.
func newRequestBuilder(client *Client, ctx context.Context, method, path string) *RequestBuilder {
	fullURL := client.baseURL + path

	return &RequestBuilder{
		client:  client,
		ctx:     ctx,
		method:  method,
		url:     fullURL,
		headers: make(http.Header),
		query:   make(url.Values),
	}
}

// Header sets a single header key-value pair.
// Returns the builder for method chaining.
func (rb *RequestBuilder) Header(key, value string) *RequestBuilder {
	rb.headers.Set(key, value)
	return rb
}

// Headers sets multiple headers from a map.
// Returns the builder for method chaining.
func (rb *RequestBuilder) Headers(headers map[string]string) *RequestBuilder {
	for key, value := range headers {
		rb.headers.Set(key, value)
	}
	return rb
}

// Query adds a single query parameter.
// Returns the builder for method chaining.
func (rb *RequestBuilder) Query(key, value string) *RequestBuilder {
	rb.query.Add(key, value)
	return rb
}

// QueryParams adds multiple query parameters from a map.
// Returns the builder for method chaining.
func (rb *RequestBuilder) QueryParams(params map[string]string) *RequestBuilder {
	for key, value := range params {
		rb.query.Add(key, value)
	}
	return rb
}

// JSON sets the request body to the JSON encoding of the provided value.
// It automatically sets the Content-Type header to application/json.
// Returns the builder for method chaining.
func (rb *RequestBuilder) JSON(v interface{}) *RequestBuilder {
	data, err := json.Marshal(v)
	if err != nil {
		// Store error to be returned by Do()
		rb.body = &errorReader{err: fmt.Errorf("encoding JSON: %w", err)}
		return rb
	}

	rb.body = bytes.NewReader(data)
	rb.headers.Set("Content-Type", "application/json")
	return rb
}

// Body sets the request body and content type.
// Returns the builder for method chaining.
func (rb *RequestBuilder) Body(body io.Reader, contentType string) *RequestBuilder {
	rb.body = body
	if contentType != "" {
		rb.headers.Set("Content-Type", contentType)
	}
	return rb
}

// Do executes the HTTP request and returns the response.
// This method should only be called once per RequestBuilder instance.
//
// Returns an error if the request cannot be built or executed.
func (rb *RequestBuilder) Do() (*Response, error) {
	// Build the full URL with query parameters
	fullURL := rb.url
	if len(rb.query) > 0 {
		fullURL = fullURL + "?" + rb.query.Encode()
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(rb.ctx, rb.method, fullURL, rb.body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Check if body had an encoding error
	if er, ok := rb.body.(*errorReader); ok {
		return nil, er.err
	}

	// Set headers
	req.Header = rb.headers

	// Execute the request through the client
	resp, err := rb.client.do(req)
	if err != nil {
		return nil, err
	}

	return &Response{Response: resp}, nil
}

// errorReader is a helper type to defer JSON encoding errors until Do() is called.
type errorReader struct {
	err error
}

// Read implements io.Reader and always returns the stored error.
func (er *errorReader) Read(p []byte) (n int, err error) {
	return 0, er.err
}

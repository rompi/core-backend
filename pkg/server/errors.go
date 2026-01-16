package server

import (
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Error represents an application error that can be returned from both gRPC and HTTP.
// It provides automatic mapping between gRPC status codes and HTTP status codes.
type Error struct {
	// Code is the gRPC status code
	Code codes.Code `json:"-"`

	// HTTPCode is the HTTP status code
	HTTPCode int `json:"code"`

	// Message is the error message
	Message string `json:"message"`

	// Details contains additional error details
	Details interface{} `json:"details,omitempty"`

	// Internal is the underlying error (not exposed to clients)
	Internal error `json:"-"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Internal)
	}
	return e.Message
}

// GRPCStatus returns the gRPC status for this error.
func (e *Error) GRPCStatus() *status.Status {
	return status.New(e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *Error) Unwrap() error {
	return e.Internal
}

// NewError creates a new Error with the given gRPC code and message.
func NewError(code codes.Code, message string) *Error {
	return &Error{
		Code:     code,
		HTTPCode: GRPCCodeToHTTP(code),
		Message:  message,
	}
}

// NewErrorWithDetails creates a new Error with details.
func NewErrorWithDetails(code codes.Code, message string, details interface{}) *Error {
	return &Error{
		Code:     code,
		HTTPCode: GRPCCodeToHTTP(code),
		Message:  message,
		Details:  details,
	}
}

// WrapError wraps an existing error with a gRPC code and message.
func WrapError(code codes.Code, message string, err error) *Error {
	return &Error{
		Code:     code,
		HTTPCode: GRPCCodeToHTTP(code),
		Message:  message,
		Internal: err,
	}
}

// FromGRPCError converts a gRPC error to our Error type.
func FromGRPCError(err error) *Error {
	if err == nil {
		return nil
	}

	// Check if it's already our Error type
	if e, ok := err.(*Error); ok {
		return e
	}

	// Try to extract gRPC status
	st, ok := status.FromError(err)
	if !ok {
		return &Error{
			Code:     codes.Unknown,
			HTTPCode: http.StatusInternalServerError,
			Message:  err.Error(),
			Internal: err,
		}
	}

	return &Error{
		Code:     st.Code(),
		HTTPCode: GRPCCodeToHTTP(st.Code()),
		Message:  st.Message(),
		Internal: err,
	}
}

// FromHTTPStatus creates an Error from an HTTP status code.
func FromHTTPStatus(httpCode int, message string) *Error {
	return &Error{
		Code:     HTTPToGRPCCode(httpCode),
		HTTPCode: httpCode,
		Message:  message,
	}
}

// Common errors for convenience.
var (
	ErrInvalidArgument    = NewError(codes.InvalidArgument, "invalid argument")
	ErrNotFound           = NewError(codes.NotFound, "not found")
	ErrAlreadyExists      = NewError(codes.AlreadyExists, "already exists")
	ErrPermissionDenied   = NewError(codes.PermissionDenied, "permission denied")
	ErrUnauthenticated    = NewError(codes.Unauthenticated, "unauthenticated")
	ErrResourceExhausted  = NewError(codes.ResourceExhausted, "rate limit exceeded")
	ErrFailedPrecondition = NewError(codes.FailedPrecondition, "failed precondition")
	ErrAborted            = NewError(codes.Aborted, "aborted")
	ErrInternal           = NewError(codes.Internal, "internal error")
	ErrUnavailable        = NewError(codes.Unavailable, "service unavailable")
	ErrDeadlineExceeded   = NewError(codes.DeadlineExceeded, "deadline exceeded")
)

// grpcToHTTP maps gRPC status codes to HTTP status codes.
var grpcToHTTP = map[codes.Code]int{
	codes.OK:                 http.StatusOK,
	codes.Canceled:           499, // Client Closed Request
	codes.Unknown:            http.StatusInternalServerError,
	codes.InvalidArgument:    http.StatusBadRequest,
	codes.DeadlineExceeded:   http.StatusGatewayTimeout,
	codes.NotFound:           http.StatusNotFound,
	codes.AlreadyExists:      http.StatusConflict,
	codes.PermissionDenied:   http.StatusForbidden,
	codes.ResourceExhausted:  http.StatusTooManyRequests,
	codes.FailedPrecondition: http.StatusBadRequest,
	codes.Aborted:            http.StatusConflict,
	codes.OutOfRange:         http.StatusBadRequest,
	codes.Unimplemented:      http.StatusNotImplemented,
	codes.Internal:           http.StatusInternalServerError,
	codes.Unavailable:        http.StatusServiceUnavailable,
	codes.DataLoss:           http.StatusInternalServerError,
	codes.Unauthenticated:    http.StatusUnauthorized,
}

// httpToGRPC maps HTTP status codes to gRPC status codes.
var httpToGRPC = map[int]codes.Code{
	http.StatusOK:                  codes.OK,
	http.StatusBadRequest:          codes.InvalidArgument,
	http.StatusUnauthorized:        codes.Unauthenticated,
	http.StatusForbidden:           codes.PermissionDenied,
	http.StatusNotFound:            codes.NotFound,
	http.StatusConflict:            codes.AlreadyExists,
	http.StatusTooManyRequests:     codes.ResourceExhausted,
	http.StatusInternalServerError: codes.Internal,
	http.StatusNotImplemented:      codes.Unimplemented,
	http.StatusServiceUnavailable:  codes.Unavailable,
	http.StatusGatewayTimeout:      codes.DeadlineExceeded,
}

// GRPCCodeToHTTP converts a gRPC status code to an HTTP status code.
func GRPCCodeToHTTP(code codes.Code) int {
	if httpCode, ok := grpcToHTTP[code]; ok {
		return httpCode
	}
	return http.StatusInternalServerError
}

// HTTPToGRPCCode converts an HTTP status code to a gRPC status code.
func HTTPToGRPCCode(httpCode int) codes.Code {
	if code, ok := httpToGRPC[httpCode]; ok {
		return code
	}

	// Default mappings for ranges
	switch {
	case httpCode >= 200 && httpCode < 300:
		return codes.OK
	case httpCode >= 400 && httpCode < 500:
		return codes.InvalidArgument
	case httpCode >= 500:
		return codes.Internal
	default:
		return codes.Unknown
	}
}

// IsNotFound returns true if the error is a not found error.
func IsNotFound(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == codes.NotFound
	}
	st, ok := status.FromError(err)
	return ok && st.Code() == codes.NotFound
}

// IsUnauthenticated returns true if the error is an unauthenticated error.
func IsUnauthenticated(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == codes.Unauthenticated
	}
	st, ok := status.FromError(err)
	return ok && st.Code() == codes.Unauthenticated
}

// IsPermissionDenied returns true if the error is a permission denied error.
func IsPermissionDenied(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == codes.PermissionDenied
	}
	st, ok := status.FromError(err)
	return ok && st.Code() == codes.PermissionDenied
}

// IsInvalidArgument returns true if the error is an invalid argument error.
func IsInvalidArgument(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == codes.InvalidArgument
	}
	st, ok := status.FromError(err)
	return ok && st.Code() == codes.InvalidArgument
}

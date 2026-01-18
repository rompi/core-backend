package server

import (
	"errors"
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name:     "message only",
			err:      &Error{Message: "something failed"},
			expected: "something failed",
		},
		{
			name: "with internal error",
			err: &Error{
				Message:  "operation failed",
				Internal: errors.New("underlying error"),
			},
			expected: "operation failed: underlying error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestError_GRPCStatus(t *testing.T) {
	err := &Error{
		Code:    codes.NotFound,
		Message: "resource not found",
	}

	st := err.GRPCStatus()
	if st.Code() != codes.NotFound {
		t.Errorf("GRPCStatus().Code() = %v, want %v", st.Code(), codes.NotFound)
	}
	if st.Message() != "resource not found" {
		t.Errorf("GRPCStatus().Message() = %q, want %q", st.Message(), "resource not found")
	}
}

func TestError_Unwrap(t *testing.T) {
	internal := errors.New("internal error")
	err := &Error{
		Message:  "wrapped",
		Internal: internal,
	}

	if err.Unwrap() != internal {
		t.Error("Unwrap() did not return internal error")
	}

	errNoInternal := &Error{Message: "no internal"}
	if errNoInternal.Unwrap() != nil {
		t.Error("Unwrap() should return nil when no internal error")
	}
}

func TestNewError(t *testing.T) {
	err := NewError(codes.InvalidArgument, "bad request")

	if err.Code != codes.InvalidArgument {
		t.Errorf("Code = %v, want %v", err.Code, codes.InvalidArgument)
	}
	if err.HTTPCode != http.StatusBadRequest {
		t.Errorf("HTTPCode = %d, want %d", err.HTTPCode, http.StatusBadRequest)
	}
	if err.Message != "bad request" {
		t.Errorf("Message = %q, want %q", err.Message, "bad request")
	}
}

func TestNewErrorWithDetails(t *testing.T) {
	details := map[string]string{"field": "email"}
	err := NewErrorWithDetails(codes.InvalidArgument, "validation error", details)

	if err.Code != codes.InvalidArgument {
		t.Errorf("Code = %v, want %v", err.Code, codes.InvalidArgument)
	}
	if err.Details == nil {
		t.Fatal("Details should not be nil")
	}
	d, ok := err.Details.(map[string]string)
	if !ok {
		t.Fatal("Details should be map[string]string")
	}
	if d["field"] != "email" {
		t.Errorf("Details[field] = %q, want %q", d["field"], "email")
	}
}

func TestWrapError(t *testing.T) {
	internal := errors.New("db connection failed")
	err := WrapError(codes.Unavailable, "service unavailable", internal)

	if err.Code != codes.Unavailable {
		t.Errorf("Code = %v, want %v", err.Code, codes.Unavailable)
	}
	if err.HTTPCode != http.StatusServiceUnavailable {
		t.Errorf("HTTPCode = %d, want %d", err.HTTPCode, http.StatusServiceUnavailable)
	}
	if err.Internal != internal {
		t.Error("Internal error not preserved")
	}
}

func TestFromGRPCError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		if FromGRPCError(nil) != nil {
			t.Error("FromGRPCError(nil) should return nil")
		}
	})

	t.Run("already our Error type", func(t *testing.T) {
		original := NewError(codes.NotFound, "not found")
		result := FromGRPCError(original)
		if result != original {
			t.Error("FromGRPCError should return same error if already our type")
		}
	})

	t.Run("gRPC status error", func(t *testing.T) {
		st := status.New(codes.PermissionDenied, "access denied")
		result := FromGRPCError(st.Err())

		if result.Code != codes.PermissionDenied {
			t.Errorf("Code = %v, want %v", result.Code, codes.PermissionDenied)
		}
		if result.Message != "access denied" {
			t.Errorf("Message = %q, want %q", result.Message, "access denied")
		}
		if result.HTTPCode != http.StatusForbidden {
			t.Errorf("HTTPCode = %d, want %d", result.HTTPCode, http.StatusForbidden)
		}
	})

	t.Run("regular error", func(t *testing.T) {
		regularErr := errors.New("some error")
		result := FromGRPCError(regularErr)

		if result.Code != codes.Unknown {
			t.Errorf("Code = %v, want %v", result.Code, codes.Unknown)
		}
		if result.HTTPCode != http.StatusInternalServerError {
			t.Errorf("HTTPCode = %d, want %d", result.HTTPCode, http.StatusInternalServerError)
		}
	})
}

func TestFromHTTPStatus(t *testing.T) {
	tests := []struct {
		httpCode int
		message  string
		wantCode codes.Code
	}{
		{http.StatusOK, "ok", codes.OK},
		{http.StatusBadRequest, "bad request", codes.InvalidArgument},
		{http.StatusUnauthorized, "unauthorized", codes.Unauthenticated},
		{http.StatusForbidden, "forbidden", codes.PermissionDenied},
		{http.StatusNotFound, "not found", codes.NotFound},
		{http.StatusConflict, "conflict", codes.AlreadyExists},
		{http.StatusTooManyRequests, "rate limited", codes.ResourceExhausted},
		{http.StatusInternalServerError, "internal", codes.Internal},
		{http.StatusNotImplemented, "not implemented", codes.Unimplemented},
		{http.StatusServiceUnavailable, "unavailable", codes.Unavailable},
		{http.StatusGatewayTimeout, "timeout", codes.DeadlineExceeded},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			err := FromHTTPStatus(tt.httpCode, tt.message)
			if err.Code != tt.wantCode {
				t.Errorf("Code = %v, want %v", err.Code, tt.wantCode)
			}
			if err.HTTPCode != tt.httpCode {
				t.Errorf("HTTPCode = %d, want %d", err.HTTPCode, tt.httpCode)
			}
			if err.Message != tt.message {
				t.Errorf("Message = %q, want %q", err.Message, tt.message)
			}
		})
	}
}

func TestGRPCCodeToHTTP(t *testing.T) {
	tests := []struct {
		code codes.Code
		want int
	}{
		{codes.OK, http.StatusOK},
		{codes.Canceled, 499},
		{codes.Unknown, http.StatusInternalServerError},
		{codes.InvalidArgument, http.StatusBadRequest},
		{codes.DeadlineExceeded, http.StatusGatewayTimeout},
		{codes.NotFound, http.StatusNotFound},
		{codes.AlreadyExists, http.StatusConflict},
		{codes.PermissionDenied, http.StatusForbidden},
		{codes.ResourceExhausted, http.StatusTooManyRequests},
		{codes.FailedPrecondition, http.StatusBadRequest},
		{codes.Aborted, http.StatusConflict},
		{codes.OutOfRange, http.StatusBadRequest},
		{codes.Unimplemented, http.StatusNotImplemented},
		{codes.Internal, http.StatusInternalServerError},
		{codes.Unavailable, http.StatusServiceUnavailable},
		{codes.DataLoss, http.StatusInternalServerError},
		{codes.Unauthenticated, http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.code.String(), func(t *testing.T) {
			if got := GRPCCodeToHTTP(tt.code); got != tt.want {
				t.Errorf("GRPCCodeToHTTP(%v) = %d, want %d", tt.code, got, tt.want)
			}
		})
	}

	t.Run("unknown code defaults to 500", func(t *testing.T) {
		if got := GRPCCodeToHTTP(codes.Code(999)); got != http.StatusInternalServerError {
			t.Errorf("GRPCCodeToHTTP(999) = %d, want %d", got, http.StatusInternalServerError)
		}
	})
}

func TestHTTPToGRPCCode(t *testing.T) {
	tests := []struct {
		httpCode int
		want     codes.Code
	}{
		{http.StatusOK, codes.OK},
		{http.StatusBadRequest, codes.InvalidArgument},
		{http.StatusUnauthorized, codes.Unauthenticated},
		{http.StatusForbidden, codes.PermissionDenied},
		{http.StatusNotFound, codes.NotFound},
		{http.StatusConflict, codes.AlreadyExists},
		{http.StatusTooManyRequests, codes.ResourceExhausted},
		{http.StatusInternalServerError, codes.Internal},
		{http.StatusNotImplemented, codes.Unimplemented},
		{http.StatusServiceUnavailable, codes.Unavailable},
		{http.StatusGatewayTimeout, codes.DeadlineExceeded},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.httpCode), func(t *testing.T) {
			if got := HTTPToGRPCCode(tt.httpCode); got != tt.want {
				t.Errorf("HTTPToGRPCCode(%d) = %v, want %v", tt.httpCode, got, tt.want)
			}
		})
	}

	t.Run("2xx defaults to OK", func(t *testing.T) {
		if got := HTTPToGRPCCode(201); got != codes.OK {
			t.Errorf("HTTPToGRPCCode(201) = %v, want %v", got, codes.OK)
		}
	})

	t.Run("4xx defaults to InvalidArgument", func(t *testing.T) {
		if got := HTTPToGRPCCode(418); got != codes.InvalidArgument {
			t.Errorf("HTTPToGRPCCode(418) = %v, want %v", got, codes.InvalidArgument)
		}
	})

	t.Run("5xx defaults to Internal", func(t *testing.T) {
		if got := HTTPToGRPCCode(502); got != codes.Internal {
			t.Errorf("HTTPToGRPCCode(502) = %v, want %v", got, codes.Internal)
		}
	})

	t.Run("other codes default to Unknown", func(t *testing.T) {
		if got := HTTPToGRPCCode(100); got != codes.Unknown {
			t.Errorf("HTTPToGRPCCode(100) = %v, want %v", got, codes.Unknown)
		}
	})
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "our Error with NotFound",
			err:  NewError(codes.NotFound, "not found"),
			want: true,
		},
		{
			name: "our Error with other code",
			err:  NewError(codes.Internal, "internal"),
			want: false,
		},
		{
			name: "gRPC status NotFound",
			err:  status.Error(codes.NotFound, "not found"),
			want: true,
		},
		{
			name: "gRPC status other",
			err:  status.Error(codes.Internal, "internal"),
			want: false,
		},
		{
			name: "regular error",
			err:  errors.New("not found"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsUnauthenticated(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "our Error with Unauthenticated",
			err:  NewError(codes.Unauthenticated, "unauthenticated"),
			want: true,
		},
		{
			name: "gRPC status Unauthenticated",
			err:  status.Error(codes.Unauthenticated, "unauthenticated"),
			want: true,
		},
		{
			name: "other error",
			err:  errors.New("error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUnauthenticated(tt.err); got != tt.want {
				t.Errorf("IsUnauthenticated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPermissionDenied(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "our Error with PermissionDenied",
			err:  NewError(codes.PermissionDenied, "denied"),
			want: true,
		},
		{
			name: "gRPC status PermissionDenied",
			err:  status.Error(codes.PermissionDenied, "denied"),
			want: true,
		},
		{
			name: "other error",
			err:  errors.New("error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPermissionDenied(tt.err); got != tt.want {
				t.Errorf("IsPermissionDenied() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsInvalidArgument(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "our Error with InvalidArgument",
			err:  NewError(codes.InvalidArgument, "invalid"),
			want: true,
		},
		{
			name: "gRPC status InvalidArgument",
			err:  status.Error(codes.InvalidArgument, "invalid"),
			want: true,
		},
		{
			name: "other error",
			err:  errors.New("error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInvalidArgument(tt.err); got != tt.want {
				t.Errorf("IsInvalidArgument() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommonErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		wantCode codes.Code
	}{
		{"ErrInvalidArgument", ErrInvalidArgument, codes.InvalidArgument},
		{"ErrNotFound", ErrNotFound, codes.NotFound},
		{"ErrAlreadyExists", ErrAlreadyExists, codes.AlreadyExists},
		{"ErrPermissionDenied", ErrPermissionDenied, codes.PermissionDenied},
		{"ErrUnauthenticated", ErrUnauthenticated, codes.Unauthenticated},
		{"ErrResourceExhausted", ErrResourceExhausted, codes.ResourceExhausted},
		{"ErrFailedPrecondition", ErrFailedPrecondition, codes.FailedPrecondition},
		{"ErrAborted", ErrAborted, codes.Aborted},
		{"ErrInternal", ErrInternal, codes.Internal},
		{"ErrUnavailable", ErrUnavailable, codes.Unavailable},
		{"ErrDeadlineExceeded", ErrDeadlineExceeded, codes.DeadlineExceeded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.wantCode {
				t.Errorf("%s.Code = %v, want %v", tt.name, tt.err.Code, tt.wantCode)
			}
		})
	}
}

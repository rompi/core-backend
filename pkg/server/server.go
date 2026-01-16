package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Server manages both gRPC and HTTP servers.
type Server struct {
	config *Config
	logger Logger

	// Addresses
	grpcAddr string
	httpAddr string

	// gRPC
	grpcServer        *grpc.Server
	grpcServerOptions []grpc.ServerOption
	unaryInterceptors []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor

	// HTTP/Gateway
	httpServer     *http.Server
	httpMux        *http.ServeMux
	gatewayMux     *runtime.ServeMux
	gatewayOptions []runtime.ServeMuxOption
	httpMiddleware []Middleware

	// Auth
	authenticator Authenticator

	// Lifecycle
	shutdownHooks []ShutdownHook
	started       bool
	mu            sync.RWMutex
}

// NewServer creates a new unified server with the given options.
func NewServer(opts ...Option) (*Server, error) {
	// Start with default config
	cfg := DefaultConfig()

	s := &Server{
		config:         cfg,
		logger:         DefaultLogger(),
		httpMux:        http.NewServeMux(),
		gatewayOptions: make([]runtime.ServeMuxOption, 0),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Set addresses from config if not set via options
	if s.grpcAddr == "" {
		s.grpcAddr = s.config.GRPCAddr()
	}
	if s.httpAddr == "" {
		s.httpAddr = s.config.HTTPAddr()
	}

	// Validate config
	if err := s.config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Initialize gateway mux with default options
	s.initGatewayMux()

	// Initialize gRPC server
	if err := s.initGRPCServer(); err != nil {
		return nil, fmt.Errorf("failed to initialize gRPC server: %w", err)
	}

	// Initialize HTTP server
	s.initHTTPServer()

	return s, nil
}

// initGatewayMux initializes the gRPC-Gateway mux with options.
func (s *Server) initGatewayMux() {
	// Add default gateway options
	defaultOpts := []runtime.ServeMuxOption{
		runtime.WithErrorHandler(s.gatewayErrorHandler),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{}),
	}

	opts := append(defaultOpts, s.gatewayOptions...)
	s.gatewayMux = runtime.NewServeMux(opts...)
}

// initGRPCServer initializes the gRPC server with interceptors.
func (s *Server) initGRPCServer() error {
	var opts []grpc.ServerOption

	// Add TLS if enabled
	if s.config.TLSEnabled {
		cert, err := tls.LoadX509KeyPair(s.config.TLSCertFile, s.config.TLSKeyFile)
		if err != nil {
			return fmt.Errorf("failed to load TLS certificates: %w", err)
		}
		creds := credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		})
		opts = append(opts, grpc.Creds(creds))
	}

	// Add interceptors
	if len(s.unaryInterceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(s.unaryInterceptors...))
	}
	if len(s.streamInterceptors) > 0 {
		opts = append(opts, grpc.ChainStreamInterceptor(s.streamInterceptors...))
	}

	// Add custom server options
	opts = append(opts, s.grpcServerOptions...)

	s.grpcServer = grpc.NewServer(opts...)
	return nil
}

// initHTTPServer initializes the HTTP server.
func (s *Server) initHTTPServer() {
	// Build the handler chain with middleware
	var handler http.Handler = s.buildHTTPHandler()

	// Apply middleware in reverse order (first middleware wraps outermost)
	for i := len(s.httpMiddleware) - 1; i >= 0; i-- {
		handler = s.httpMiddleware[i](handler)
	}

	s.httpServer = &http.Server{
		Addr:         s.httpAddr,
		Handler:      handler,
		ReadTimeout:  s.config.HTTPReadTimeout,
		WriteTimeout: s.config.HTTPWriteTimeout,
		IdleTimeout:  s.config.HTTPIdleTimeout,
	}

	// Configure TLS if enabled
	if s.config.TLSEnabled {
		s.httpServer.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}
}

// buildHTTPHandler builds the combined HTTP handler.
func (s *Server) buildHTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check custom handlers first
		if handler, pattern := s.httpMux.Handler(r); pattern != "" {
			handler.ServeHTTP(w, r)
			return
		}

		// Fall back to gateway
		s.gatewayMux.ServeHTTP(w, r)
	})
}

// gatewayErrorHandler handles errors from the gRPC-Gateway.
func (s *Server) gatewayErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	// Convert gRPC error to our error type
	appErr := FromGRPCError(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.HTTPCode)

	// Write error response
	resp := map[string]interface{}{
		"code":    appErr.HTTPCode,
		"message": appErr.Message,
	}
	if appErr.Details != nil {
		resp["details"] = appErr.Details
	}

	if encErr := marshaler.NewEncoder(w).Encode(resp); encErr != nil {
		s.logger.Error("failed to encode error response", "error", encErr)
	}
}

// --- gRPC Registration ---

// GRPCServer returns the underlying gRPC server for service registration.
func (s *Server) GRPCServer() *grpc.Server {
	return s.grpcServer
}

// RegisterService is a helper to register a gRPC service.
func (s *Server) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.grpcServer.RegisterService(desc, impl)
}

// --- Gateway Registration ---

// GatewayMux returns the gRPC-Gateway mux for custom registration.
func (s *Server) GatewayMux() *runtime.ServeMux {
	return s.gatewayMux
}

// RegisterGateway registers a gRPC-Gateway handler using an endpoint connection.
// The handler is generated by protoc-gen-grpc-gateway.
// Example: server.RegisterGateway(ctx, pb.RegisterUserServiceHandlerFromEndpoint)
func (s *Server) RegisterGateway(ctx context.Context, registerFunc func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error) error {
	var opts []grpc.DialOption
	if s.config.TLSEnabled {
		// Use TLS credentials for gateway-to-gRPC connection
		creds, err := credentials.NewClientTLSFromFile(s.config.TLSCertFile, "")
		if err != nil {
			return fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	return registerFunc(ctx, s.gatewayMux, s.grpcAddr, opts)
}

// RegisterGatewayHandler registers a gateway handler directly (for in-process).
// Example: server.RegisterGatewayHandler(pb.RegisterUserServiceHandler, userServiceImpl)
func (s *Server) RegisterGatewayHandler(registerFunc func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error, conn *grpc.ClientConn) error {
	return registerFunc(context.Background(), s.gatewayMux, conn)
}

// --- Custom HTTP Handlers ---

// HandleFunc registers a custom HTTP handler function.
// Example: server.HandleFunc("/webhooks/stripe", stripeHandler)
func (s *Server) HandleFunc(pattern string, handler http.HandlerFunc) {
	s.httpMux.HandleFunc(pattern, handler)
}

// Handle registers a custom HTTP handler.
func (s *Server) Handle(pattern string, handler http.Handler) {
	s.httpMux.Handle(pattern, handler)
}

// --- Lifecycle ---

// Start starts both gRPC and HTTP servers (non-blocking).
func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return fmt.Errorf("server already started")
	}

	// Start gRPC server
	grpcLis, err := net.Listen("tcp", s.grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.grpcAddr, err)
	}

	go func() {
		s.logger.Info("gRPC server starting", "addr", s.grpcAddr)
		if err := s.grpcServer.Serve(grpcLis); err != nil {
			s.logger.Error("gRPC server error", "error", err)
		}
	}()

	// Start HTTP server
	go func() {
		s.logger.Info("HTTP server starting", "addr", s.httpAddr)
		var err error
		if s.config.TLSEnabled {
			err = s.httpServer.ListenAndServeTLS(s.config.TLSCertFile, s.config.TLSKeyFile)
		} else {
			err = s.httpServer.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server error", "error", err)
		}
	}()

	s.started = true
	return nil
}

// ListenAndServe starts the servers and blocks until shutdown.
// Handles OS signals (SIGINT, SIGTERM) for graceful shutdown.
func (s *Server) ListenAndServe() error {
	if err := s.Start(); err != nil {
		return err
	}

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Shutting down servers...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	return s.Shutdown(ctx)
}

// Shutdown gracefully shuts down both servers.
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var errs []error

	// Run shutdown hooks
	for _, hook := range s.shutdownHooks {
		if err := hook(); err != nil {
			errs = append(errs, fmt.Errorf("shutdown hook error: %w", err))
		}
	}

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("HTTP server shutdown error: %w", err))
	}

	// Graceful stop gRPC server
	done := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-ctx.Done():
		s.grpcServer.Stop() // Force stop if context expires
		errs = append(errs, fmt.Errorf("gRPC server shutdown timeout"))
	case <-done:
		// Graceful shutdown completed
	}

	s.started = false

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	s.logger.Info("Servers stopped")
	return nil
}

// OnShutdown registers a hook to be called during shutdown.
func (s *Server) OnShutdown(hook ShutdownHook) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shutdownHooks = append(s.shutdownHooks, hook)
}

// --- Accessors ---

// GRPCAddr returns the gRPC server address.
func (s *Server) GRPCAddr() string {
	return s.grpcAddr
}

// HTTPAddr returns the HTTP server address.
func (s *Server) HTTPAddr() string {
	return s.httpAddr
}

// Config returns the server configuration.
func (s *Server) Config() *Config {
	return s.config
}

// Logger returns the server logger.
func (s *Server) Logger() Logger {
	return s.logger
}

// Authenticator returns the server authenticator.
func (s *Server) Authenticator() Authenticator {
	return s.authenticator
}

// IsStarted returns whether the server has been started.
func (s *Server) IsStarted() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.started
}

// Wait blocks until the server receives a shutdown signal.
func (s *Server) Wait() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

// MustStart starts the server and panics on error.
func (s *Server) MustStart() {
	if err := s.Start(); err != nil {
		panic(err)
	}
}

// ServeHTTP implements http.Handler for testing purposes.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.httpServer.Handler.ServeHTTP(w, r)
}

// DialGRPC creates a gRPC client connection to this server.
// Useful for in-process testing.
func (s *Server) DialGRPC(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if s.config.TLSEnabled {
		creds, err := credentials.NewClientTLSFromFile(s.config.TLSCertFile, "")
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	return grpc.DialContext(ctx, s.grpcAddr, opts...)
}

// WaitForReady waits for the server to be ready to accept connections.
func (s *Server) WaitForReady(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Try to connect to HTTP server
			resp, err := http.Get(fmt.Sprintf("http://%s%s", s.httpAddr, s.config.HealthHTTPPath))
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return nil
				}
			}
		}
	}
}

package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/cloudflare/tableflip"
	"medical-rep/configs"
	"medical-rep/internal/platform/logger"
)

// Server represents the HTTP server
type Server struct {
	config   *configs.Config
	logger   *logger.Logger
	server   *http.Server
	upgrader *tableflip.Upgrader
	listener net.Listener
}

// ServerOptions holds server configuration options
type ServerOptions struct {
	Config   *configs.Config
	Logger   *logger.Logger
	Handler  http.Handler
	Upgrader *tableflip.Upgrader
}

// NewServer creates a new HTTP server instance
func NewServer(opts ServerOptions) (*Server, error) {
	if opts.Config == nil {
		return nil, fmt.Errorf("config is required")
	}
	if opts.Logger == nil {
		return nil, fmt.Errorf("logger is required")
	}
	if opts.Handler == nil {
		return nil, fmt.Errorf("handler is required")
	}
	if opts.Upgrader == nil {
		return nil, fmt.Errorf("upgrader is required")
	}

	s := &Server{
		config:   opts.Config,
		logger:   opts.Logger,
		upgrader: opts.Upgrader,
	}

	if err := s.setupServer(opts.Handler); err != nil {
		return nil, fmt.Errorf("failed to setup server: %w", err)
	}

	return s, nil
}

// setupServer configures the HTTP server with proper settings
func (s *Server) setupServer(handler http.Handler) error {
	addr := fmt.Sprintf("%s:%d", s.config.HTTP.Host, s.config.HTTP.Port)

	s.server = &http.Server{
		Addr:           addr,
		Handler:        handler,
		ReadTimeout:    s.config.HTTP.ReadTimeout,
		WriteTimeout:   s.config.HTTP.WriteTimeout,
		IdleTimeout:    s.config.HTTP.IdleTimeout,
		MaxHeaderBytes: s.config.HTTP.MaxHeaderBytes,
		ErrorLog:       s.logger.StdLogger(),
	}

	// Configure TLS if enabled
	if s.config.HTTP.TLS.Enabled {
		tlsConfig, err := s.setupTLS()
		if err != nil {
			return fmt.Errorf("failed to setup TLS: %w", err)
		}
		s.server.TLSConfig = tlsConfig
	}

	return nil
}

// setupTLS configures TLS settings
func (s *Server) setupTLS() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(s.config.HTTP.TLS.CertFile, s.config.HTTP.TLS.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		},
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519,
		},
	}

	return tlsConfig, nil
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Create listener using tableflip for zero-downtime deployments
	addr := fmt.Sprintf("%s:%d", s.config.HTTP.Host, s.config.HTTP.Port)
	ln, err := s.upgrader.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.listener = ln

	s.logger.Info("HTTP server starting",
		"addr", addr,
		"tls_enabled", s.config.HTTP.TLS.Enabled,
		"environment", s.config.App.Environment,
	)

	// Start server
	if s.config.HTTP.TLS.Enabled {
		return s.server.ServeTLS(ln, "", "")
	}

	return s.server.Serve(ln)
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server...")

	// Shutdown server gracefully
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("Server shutdown error", "error", err)
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	s.logger.Info("HTTP server stopped successfully")
	return nil
}

// Addr returns the server's network address
func (s *Server) Addr() net.Addr {
	if s.listener != nil {
		return s.listener.Addr()
	}
	return nil
}

// IsReady checks if the server is ready to serve requests
func (s *Server) IsReady() bool {
	return s.listener != nil
}

// HealthCheck performs a health check on the server
func (s *Server) HealthCheck(ctx context.Context) error {
	if !s.IsReady() {
		return fmt.Errorf("server is not ready")
	}

	// Simple health check - attempt to connect to our own address
	addr := s.Addr()
	if addr == nil {
		return fmt.Errorf("server address not available")
	}

	conn, err := net.DialTimeout("tcp", addr.String(), 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()

	return nil
}

// GetMetrics returns server metrics (placeholder for future implementation)
func (s *Server) GetMetrics() map[string]interface{} {
	metrics := map[string]interface{}{
		"server_ready": s.IsReady(),
		"tls_enabled":  s.config.HTTP.TLS.Enabled,
		"port":         s.config.HTTP.Port,
		"host":         s.config.HTTP.Host,
	}

	if s.listener != nil {
		metrics["addr"] = s.listener.Addr().String()
	}

	return metrics
}
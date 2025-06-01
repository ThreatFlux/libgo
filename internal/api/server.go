package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/threatflux/libgo/internal/config"
	"github.com/threatflux/libgo/pkg/logger"
)

// Server represents the HTTP server
type Server struct {
	router     *gin.Engine
	httpServer *http.Server
	config     config.ServerConfig
	logger     logger.Logger
	certFile   string
	keyFile    string
}

// NewServer creates a new API server
func NewServer(config config.ServerConfig, logger logger.Logger) *Server {
	// Set Gin mode
	if config.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else if config.Mode == "test" {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create router
	router := gin.New()

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	httpServer := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		MaxHeaderBytes: config.MaxHeaderBytes,
	}

	return &Server{
		router:     router,
		httpServer: httpServer,
		config:     config,
		logger:     logger,
		certFile:   config.TLS.CertFile,
		keyFile:    config.TLS.KeyFile,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("Starting API server",
		logger.String("address", s.httpServer.Addr),
		logger.Bool("tls", s.config.TLS.Enabled))

	// Start the server
	if s.config.TLS.Enabled {
		return s.httpServer.ListenAndServeTLS(s.certFile, s.keyFile)
	}
	return s.httpServer.ListenAndServe()
}

// Stop stops the HTTP server gracefully
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping API server")
	return s.httpServer.Shutdown(ctx)
}

// Router returns the Gin router
func (s *Server) Router() *gin.Engine {
	return s.router
}

// ConfigureMiddleware configures the middleware for the router
func (s *Server) ConfigureMiddleware(middlewares ...gin.HandlerFunc) {
	s.router.Use(middlewares...)
}

// Address returns the server's address
func (s *Server) Address() string {
	return s.httpServer.Addr
}

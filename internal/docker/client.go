package docker

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/threatflux/libgo/pkg/logger"
)

// Common errors with detailed descriptions for better error handling.
var (
	ErrNilOption              = errors.New("nil option provided to client configuration")
	ErrInvalidHost            = errors.New("invalid Docker host specification")
	ErrMissingTLSConfig       = errors.New("TLS verification enabled but certificate paths not provided")
	ErrInvalidTLSCert         = errors.New("invalid or inaccessible TLS certificate")
	ErrInvalidTLSKey          = errors.New("invalid or inaccessible TLS key")
	ErrInvalidTLSCA           = errors.New("invalid or inaccessible TLS CA certificate")
	ErrConnectionFailed       = errors.New("failed to connect to Docker daemon")
	ErrClientNotInitialized   = errors.New("Docker client not initialized")
	ErrClientClosed           = errors.New("Docker client manager has been closed")
	ErrInvalidAPIVersion      = errors.New("invalid Docker API version format")
	ErrContextCanceled        = errors.New("context was canceled while operating Docker client")
	ErrEmptyOption            = errors.New("empty value provided for required option")
	ErrTLSConfigValidation    = errors.New("TLS configuration validation failed")
	ErrCertificateExpired     = errors.New("TLS certificate has expired")
	ErrCertificateNotYetValid = errors.New("TLS certificate is not yet valid")
)

// ClientOption represents a functional option for configuring the Docker client.
type ClientOption func(*ClientConfig) error

// ClientConfig represents the configuration for the Docker client.
type ClientConfig struct {
	Logger                      logger.Logger
	Headers                     map[string]string
	DialContext                 func(ctx context.Context, network, addr string) (net.Conn, error)
	Host                        string
	APIVersion                  string
	TLSCertPath                 string
	TLSKeyPath                  string
	TLSCAPath                   string
	TLSCipherSuites             []uint16
	KeepAlive                   time.Duration
	MaxIdleConns                int
	TLSHandshakeTimeout         time.Duration
	ConnectionTimeout           time.Duration
	IdleConnTimeout             time.Duration
	ResponseHeaderTimeout       time.Duration
	ExpectContinueTimeout       time.Duration
	PingTimeout                 time.Duration
	RetryDelay                  time.Duration
	ConnectionIdleTimeout       time.Duration
	MaxIdleConnsPerHost         int
	MaxConnsPerHost             int
	RetryCount                  int
	RequestTimeout              time.Duration
	TLSMinVersion               uint16
	TLSMaxVersion               uint16
	TLSVerify                   bool
	TLSPreferServerCipherSuites bool
}

// ClientManager manages Docker clients.
type ClientManager struct {
	lastPing    time.Time
	logger      logger.Logger
	client      *client.Client
	config      ClientConfig
	createCount int64
	mu          sync.RWMutex
	clientMu    sync.Mutex
	pingMutex   sync.Mutex
	initialized atomic.Bool
	closed      bool
}

// lockedClientWrapper wraps a Docker client to ensure thread-safe access.
type lockedClientWrapper struct {
	client.APIClient
	mu *sync.Mutex
}

// DefaultClientConfig returns the default client configuration.
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		Host:                  "unix:///var/run/docker.sock",
		APIVersion:            "",
		TLSVerify:             false,
		RequestTimeout:        30 * time.Second,
		ConnectionTimeout:     15 * time.Second,
		ConnectionIdleTimeout: 60 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		KeepAlive:             30 * time.Second,
		MaxIdleConns:          10,
		MaxIdleConnsPerHost:   5,
		MaxConnsPerHost:       20,
		IdleConnTimeout:       90 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		Headers:               make(map[string]string),
		PingTimeout:           5 * time.Second,
		RetryCount:            3,
		RetryDelay:            500 * time.Millisecond,
		TLSMinVersion:         tls.VersionTLS12,
		TLSMaxVersion:         0,
		TLSCipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		TLSPreferServerCipherSuites: true,
		DialContext: (&net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}
}

// WithHost sets the Docker daemon host.
func WithHost(host string) ClientOption {
	return func(config *ClientConfig) error {
		if host == "" {
			return ErrInvalidHost
		}

		if !strings.HasPrefix(host, "unix://") && !strings.HasPrefix(host, "tcp://") &&
			!strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
			return fmt.Errorf("%w: host must start with unix://, tcp://, http://, or https://", ErrInvalidHost)
		}

		config.Host = host
		return nil
	}
}

// WithAPIVersion sets the Docker API version.
func WithAPIVersion(version string) ClientOption {
	return func(config *ClientConfig) error {
		if version == "" {
			config.APIVersion = ""
			return nil
		}

		if !strings.HasPrefix(version, "v") {
			parts := strings.Split(version, ".")
			if len(parts) != 2 {
				return fmt.Errorf("%w: version should be in format vX.Y or X.Y", ErrInvalidAPIVersion)
			}
		}

		config.APIVersion = version
		return nil
	}
}

// WithTLSVerify enables TLS verification.
func WithTLSVerify(verify bool) ClientOption {
	return func(config *ClientConfig) error {
		config.TLSVerify = verify
		return nil
	}
}

// WithTLSConfig sets the complete TLS configuration.
func WithTLSConfig(certPath, keyPath, caPath string) ClientOption {
	return func(config *ClientConfig) error {
		if certPath == "" || keyPath == "" || caPath == "" {
			return ErrMissingTLSConfig
		}

		config.TLSVerify = true
		config.TLSCertPath = certPath
		config.TLSKeyPath = keyPath
		config.TLSCAPath = caPath

		return nil
	}
}

// WithLogger sets the logger.
func WithLogger(logger logger.Logger) ClientOption {
	return func(config *ClientConfig) error {
		if logger == nil {
			return fmt.Errorf("logger cannot be nil")
		}
		config.Logger = logger
		return nil
	}
}

// WithRequestTimeout sets the request timeout.
func WithRequestTimeout(timeout time.Duration) ClientOption {
	return func(config *ClientConfig) error {
		if timeout <= 0 {
			return fmt.Errorf("request timeout must be positive")
		}
		config.RequestTimeout = timeout
		return nil
	}
}

// WithRetry sets retry parameters.
func WithRetry(count int, delay time.Duration) ClientOption {
	return func(config *ClientConfig) error {
		if count < 0 {
			return fmt.Errorf("retry count must be non-negative")
		}
		if delay < 0 {
			return fmt.Errorf("retry delay must be non-negative")
		}

		config.RetryCount = count
		config.RetryDelay = delay
		return nil
	}
}

// NewManager creates a new Docker client manager.
func NewManager(opts ...ClientOption) (*ClientManager, error) {
	config := DefaultClientConfig()

	for _, opt := range opts {
		if opt == nil {
			return nil, ErrNilOption
		}
		if err := opt(&config); err != nil {
			return nil, fmt.Errorf("option application failed: %w", err)
		}
	}

	if config.TLSVerify {
		if config.TLSCertPath == "" || config.TLSKeyPath == "" || config.TLSCAPath == "" {
			return nil, ErrMissingTLSConfig
		}
	}

	manager := &ClientManager{
		config: config,
		logger: config.Logger,
		closed: false,
	}

	_, err := manager.GetClient()
	if err != nil {
		if manager.logger != nil {
			manager.logger.Warn("Initial Docker client creation failed", logger.Error(err))
		}
		manager.initialized.Store(false)
	} else {
		manager.initialized.Store(true)
	}

	return manager, nil
}

// GetClient returns a thread-safe Docker API client wrapper.
func (m *ClientManager) GetClient() (client.APIClient, error) {
	return m.GetWithContext(context.Background())
}

// GetWithContext returns a thread-safe Docker API client wrapper with context.
func (m *ClientManager) GetWithContext(ctx context.Context) (client.APIClient, error) {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return nil, ErrClientClosed
	}
	if m.client != nil {
		pingCtx, cancel := context.WithTimeout(ctx, m.config.PingTimeout)
		defer cancel()

		_, err := m.client.Ping(pingCtx)
		if err == nil {
			m.mu.RUnlock()
			return &lockedClientWrapper{APIClient: m.client, mu: &m.clientMu}, nil
		}
		if m.logger != nil {
			m.logger.Warn("Existing Docker client failed ping", logger.Error(err))
		}
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil, ErrClientClosed
	}
	if m.client != nil {
		pingCtx, cancel := context.WithTimeout(ctx, m.config.PingTimeout)
		defer cancel()
		_, err := m.client.Ping(pingCtx)
		if err == nil {
			return &lockedClientWrapper{APIClient: m.client, mu: &m.clientMu}, nil
		}
		if m.logger != nil {
			m.logger.Warn("Existing Docker client still failing ping after acquiring lock", logger.Error(err))
		}
	}

	var newClient *client.Client
	var lastErr error
	for i := 0; i <= m.config.RetryCount; i++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w: %w", ErrContextCanceled, ctx.Err())
		default:
		}

		if m.logger != nil {
			m.logger.Debug("Attempting to create Docker client", logger.Int("attempt", i+1), logger.Int("max_attempts", m.config.RetryCount+1))
		}
		newClient, lastErr = m.createClient(ctx)
		if lastErr == nil {
			if m.logger != nil {
				m.logger.Info("Successfully created Docker client", logger.Int("attempt", i+1))
			}
			m.client = newClient
			m.initialized.Store(true)
			m.createCount++
			return &lockedClientWrapper{APIClient: m.client, mu: &m.clientMu}, nil
		}

		if m.logger != nil {
			m.logger.Warn("Error creating Docker client", logger.Int("attempt", i+1), logger.Error(lastErr))
		}
		if i < m.config.RetryCount {
			select {
			case <-time.After(m.config.RetryDelay):
			case <-ctx.Done():
				return nil, fmt.Errorf("%w during retry delay: %w", ErrContextCanceled, ctx.Err())
			}
		}
	}

	m.initialized.Store(false)
	return nil, fmt.Errorf("failed to create Docker client after %d attempts: %w", m.config.RetryCount+1, lastErr)
}

// createClient handles the actual client creation logic.
func (m *ClientManager) createClient(ctx context.Context) (*client.Client, error) {
	var opts []client.Opt

	if m.config.Host != "" {
		opts = append(opts, client.WithHost(m.config.Host))
	} else {
		opts = append(opts, client.FromEnv)
	}

	if m.config.APIVersion != "" {
		opts = append(opts, client.WithVersion(m.config.APIVersion))
	} else {
		opts = append(opts, client.WithAPIVersionNegotiation())
	}

	isUnixSocket := strings.HasPrefix(m.config.Host, "unix://")
	if !isUnixSocket || m.config.TLSVerify {
		httpClient := m.createSecureHTTPClient()
		if httpClient != nil {
			opts = append(opts, client.WithHTTPClient(httpClient))
		} else {
			if m.logger != nil {
				m.logger.Error("Failed to create secure HTTP client when required")
			}
		}
	}

	if len(m.config.Headers) > 0 {
		opts = append(opts, client.WithHTTPHeaders(m.config.Headers))
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrConnectionFailed, err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, m.config.PingTimeout)
	defer cancel()
	_, err = cli.Ping(pingCtx)
	if err != nil {
		cli.Close()
		return nil, fmt.Errorf("failed to ping Docker daemon after connection: %w", err)
	}

	if m.logger != nil {
		m.logger.Debug("Docker client created and ping successful")
	}
	return cli, nil
}

// createSecureHTTPClient creates an *http.Client with TLS and timeout settings.
func (m *ClientManager) createSecureHTTPClient() *http.Client {
	transport := m.createHTTPTransport()
	m.configureTLSForTransport(transport)

	return &http.Client{
		Transport: transport,
		Timeout:   m.config.RequestTimeout,
	}
}

// createHTTPTransport creates the base HTTP transport with connection settings.
func (m *ClientManager) createHTTPTransport() *http.Transport {
	return &http.Transport{
		Proxy:       http.ProxyFromEnvironment,
		DialContext: m.createDialContextFunc(),
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          m.config.MaxIdleConns,
		IdleConnTimeout:       m.config.IdleConnTimeout,
		TLSHandshakeTimeout:   m.config.TLSHandshakeTimeout,
		ExpectContinueTimeout: m.config.ExpectContinueTimeout,
		MaxIdleConnsPerHost:   m.config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       m.config.MaxConnsPerHost,
		ResponseHeaderTimeout: m.config.ResponseHeaderTimeout,
	}
}

// createDialContextFunc creates the dial context function for the transport.
func (m *ClientManager) createDialContextFunc() func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialerFunc := m.config.DialContext
		if dialerFunc == nil {
			dialerFunc = (&net.Dialer{
				Timeout:   m.config.ConnectionTimeout,
				KeepAlive: m.config.KeepAlive,
			}).DialContext
		}

		if strings.HasPrefix(m.config.Host, "unix://") {
			socketPath := strings.TrimPrefix(m.config.Host, "unix://")
			return dialerFunc(ctx, "unix", socketPath)
		}

		return dialerFunc(ctx, network, addr)
	}
}

// configureTLSForTransport configures TLS settings for the HTTP transport.
func (m *ClientManager) configureTLSForTransport(transport *http.Transport) {
	if m.config.TLSVerify {
		tlsConfig := m.createTLSConfig()
		m.loadTLSCertificates(tlsConfig, transport)
		m.loadTLSCACertificate(tlsConfig, transport)
		transport.TLSClientConfig = tlsConfig
	} else {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
}

// createTLSConfig creates the base TLS configuration.
func (m *ClientManager) createTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion:               m.config.TLSMinVersion,
		MaxVersion:               m.config.TLSMaxVersion,
		CipherSuites:             m.config.TLSCipherSuites,
		PreferServerCipherSuites: m.config.TLSPreferServerCipherSuites,
	}
}

// loadTLSCertificates loads client certificates for mutual TLS.
func (m *ClientManager) loadTLSCertificates(tlsConfig *tls.Config, transport *http.Transport) {
	if m.config.TLSCertPath != "" && m.config.TLSKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(m.config.TLSCertPath, m.config.TLSKeyPath)
		if err != nil {
			if m.logger != nil {
				m.logger.Error("Failed to load TLS key pair", logger.Error(err))
			}
			return
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
}

// loadTLSCACertificate loads CA certificate for server verification.
func (m *ClientManager) loadTLSCACertificate(tlsConfig *tls.Config, transport *http.Transport) {
	if m.config.TLSCAPath != "" {
		caCert, err := os.ReadFile(m.config.TLSCAPath)
		if err != nil {
			if m.logger != nil {
				m.logger.Error("Failed to read CA certificate", logger.Error(err))
			}
			return
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			if m.logger != nil {
				m.logger.Error("Failed to append CA certificate to pool")
			}
			return
		}
		tlsConfig.RootCAs = caCertPool
	}
}

// Ping checks the connectivity with the Docker daemon.
func (m *ClientManager) Ping(ctx context.Context) (types.Ping, error) {
	m.pingMutex.Lock()
	defer m.pingMutex.Unlock()

	cli, err := m.GetWithContext(ctx)
	if err != nil {
		return types.Ping{}, fmt.Errorf("failed to get Docker client for ping: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, m.config.PingTimeout)
	defer cancel()

	pingResult, err := cli.Ping(pingCtx)
	if err != nil {
		if m.logger != nil {
			m.logger.Error("Ping failed", logger.Error(err))
		}
		return types.Ping{}, fmt.Errorf("Docker daemon ping failed: %w", err)
	}

	m.lastPing = time.Now()
	if m.logger != nil {
		m.logger.Debug("Ping successful",
			logger.String("api_version", pingResult.APIVersion),
			logger.String("os_type", pingResult.OSType),
			logger.Bool("experimental", pingResult.Experimental))
	}

	return pingResult, nil
}

// Close closes the managed Docker client.
func (m *ClientManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		if m.logger != nil {
			m.logger.Debug("Client manager already closed")
		}
		return nil
	}

	m.closed = true
	m.initialized.Store(false)

	if m.client != nil {
		if m.logger != nil {
			m.logger.Info("Closing Docker client")
		}
		err := m.client.Close()
		m.client = nil
		if err != nil {
			if m.logger != nil {
				m.logger.Error("Error closing Docker client", logger.Error(err))
			}
			return fmt.Errorf("failed to close Docker client: %w", err)
		}
		if m.logger != nil {
			m.logger.Info("Docker client closed successfully")
		}
		return nil
	}

	if m.logger != nil {
		m.logger.Info("No active Docker client to close")
	}
	return nil
}

// IsInitialized checks if the client manager has successfully initialized.
func (m *ClientManager) IsInitialized() bool {
	return m.initialized.Load()
}

// IsClosed checks if the client manager has been closed.
func (m *ClientManager) IsClosed() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.closed
}

// GetConfig returns a copy of the current client configuration.
func (m *ClientManager) GetConfig() ClientConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	configCopy := m.config
	if configCopy.Headers != nil {
		configCopy.Headers = make(map[string]string)
		for k, v := range m.config.Headers {
			configCopy.Headers[k] = v
		}
	}
	return configCopy
}

// Wrapper methods for thread-safe access - only implementing essential ones for brevity.
func (w *lockedClientWrapper) ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.APIClient.ContainerList(ctx, options)
}

func (w *lockedClientWrapper) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.APIClient.ContainerCreate(ctx, config, hostConfig, networkingConfig, platform, containerName)
}

func (w *lockedClientWrapper) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.APIClient.ContainerStart(ctx, containerID, options)
}

func (w *lockedClientWrapper) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.APIClient.ContainerStop(ctx, containerID, options)
}

func (w *lockedClientWrapper) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.APIClient.ContainerRemove(ctx, containerID, options)
}

func (w *lockedClientWrapper) ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.APIClient.ContainerInspect(ctx, containerID)
}

func (w *lockedClientWrapper) Ping(ctx context.Context) (types.Ping, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.APIClient.Ping(ctx)
}

func (w *lockedClientWrapper) ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.APIClient.ImageList(ctx, options)
}

func (w *lockedClientWrapper) NetworkList(ctx context.Context, options network.ListOptions) ([]network.Summary, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.APIClient.NetworkList(ctx, options)
}

func (w *lockedClientWrapper) VolumeList(ctx context.Context, options volume.ListOptions) (volume.ListResponse, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.APIClient.VolumeList(ctx, options)
}

// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package contracts

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
)

// MockLogger is a simple logger implementation for testing
type MockLogger struct {
	DebugMessages []string
	InfoMessages  []string
	WarnMessages  []string
	ErrorMessages []string
	mu            sync.Mutex
}

// NewMockLogger creates a new mock logger
func NewMockLogger() *MockLogger {
	return &MockLogger{
		DebugMessages: []string{},
		InfoMessages:  []string{},
		WarnMessages:  []string{},
		ErrorMessages: []string{},
	}
}

// Debug logs a debug message
func (l *MockLogger) Debug(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.DebugMessages = append(l.DebugMessages, fmt.Sprintf(format, args...))
}

// Info logs an info message
func (l *MockLogger) Info(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.InfoMessages = append(l.InfoMessages, fmt.Sprintf(format, args...))
}

// Warn logs a warning message
func (l *MockLogger) Warn(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.WarnMessages = append(l.WarnMessages, fmt.Sprintf(format, args...))
}

// Error logs an error message
func (l *MockLogger) Error(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ErrorMessages = append(l.ErrorMessages, fmt.Sprintf(format, args...))
}

// MockClient is a simple client implementation for testing
type MockClient struct {
	BaseURL      string
	HTTPClient   *http.Client
	UserAgent    string
	Logger       interfaces.Logger
	ResponseFunc func(req *http.Request) (*http.Response, error)
}

// NewMockClient creates a new mock client
func NewMockClient() *MockClient {
	return &MockClient{
		BaseURL:    "https://example.com",
		HTTPClient: &http.Client{},
		UserAgent:  "mock-client/1.0",
		Logger:     NewMockLogger(),
		ResponseFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"status": "ok"}`)),
				Header:     make(http.Header),
			}, nil
		},
	}
}

// Do performs an HTTP request
func (c *MockClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// continue
	}

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if c.ResponseFunc != nil {
		return c.ResponseFunc(req)
	}

	return c.HTTPClient.Do(req)
}

// GetHTTPClient returns the HTTP client
func (c *MockClient) GetHTTPClient() *http.Client {
	return c.HTTPClient
}

// GetBaseURL returns the base URL
func (c *MockClient) GetBaseURL() string {
	return c.BaseURL
}

// GetUserAgent returns the user agent
func (c *MockClient) GetUserAgent() string {
	return c.UserAgent
}

// GetLogger returns the logger
func (c *MockClient) GetLogger() interfaces.Logger {
	return c.Logger
}

// MockTransport is a simple transport implementation for testing
type MockTransport struct {
	Logger       interfaces.Logger
	BaseURL      string
	ResponseFunc func(req *http.Request) (*http.Response, error)
}

// NewMockTransport creates a new mock transport
func NewMockTransport() *MockTransport {
	return &MockTransport{
		Logger:  NewMockLogger(),
		BaseURL: "https://example.com",
		ResponseFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"status": "ok"}`)),
				Header:     make(http.Header),
			}, nil
		},
	}
}

// Request makes a request with the given method, path, body, query, and headers
func (t *MockTransport) Request(ctx context.Context, method, path string, body interface{},
	query url.Values, headers http.Header) (*http.Response, error) {

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// continue
	}

	// Validate method
	if method == "" {
		return nil, fmt.Errorf("method cannot be empty")
	}

	// Check for invalid methods
	validMethods := map[string]bool{
		http.MethodGet:     true,
		http.MethodPost:    true,
		http.MethodPut:     true,
		http.MethodDelete:  true,
		http.MethodHead:    true,
		http.MethodPatch:   true,
		http.MethodOptions: true,
	}

	if !validMethods[method] {
		return nil, fmt.Errorf("invalid method: %s", method)
	}

	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	var bodyReader io.Reader
	if body != nil {
		if r, ok := body.(io.Reader); ok {
			bodyReader = r
		} else {
			bodyReader = strings.NewReader(fmt.Sprintf("%v", body))
		}
	}

	u, err := url.Parse(t.BaseURL + path)
	if err != nil {
		return nil, err
	}

	if query != nil {
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequest(method, u.String(), bodyReader)
	if err != nil {
		return nil, err
	}

	if headers != nil {
		req.Header = headers
	}

	if t.ResponseFunc != nil {
		return t.ResponseFunc(req)
	}

	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{"status": "ok"}`)),
		Header:     make(http.Header),
	}, nil
}

// Get makes a GET request
func (t *MockTransport) Get(ctx context.Context, path string, query url.Values, headers http.Header) (*http.Response, error) {
	return t.Request(ctx, "GET", path, nil, query, headers)
}

// Post makes a POST request
func (t *MockTransport) Post(ctx context.Context, path string, body interface{}, query url.Values, headers http.Header) (*http.Response, error) {
	return t.Request(ctx, "POST", path, body, query, headers)
}

// Put makes a PUT request
func (t *MockTransport) Put(ctx context.Context, path string, body interface{}, query url.Values, headers http.Header) (*http.Response, error) {
	return t.Request(ctx, "PUT", path, body, query, headers)
}

// Delete makes a DELETE request
func (t *MockTransport) Delete(ctx context.Context, path string, query url.Values, headers http.Header) (*http.Response, error) {
	return t.Request(ctx, "DELETE", path, nil, query, headers)
}

// Patch makes a PATCH request
func (t *MockTransport) Patch(ctx context.Context, path string, body interface{}, query url.Values, headers http.Header) (*http.Response, error) {
	return t.Request(ctx, "PATCH", path, body, query, headers)
}

// RoundTrip implements http.RoundTripper
func (t *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if t.ResponseFunc != nil {
		return t.ResponseFunc(req)
	}

	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{"status": "ok"}`)),
		Header:     make(http.Header),
	}, nil
}

// MockAuthorizer is a simple authorizer implementation for testing
type MockAuthorizer struct {
	Token  string
	Valid  bool
	Logger interfaces.Logger
}

// NewMockAuthorizer creates a new mock authorizer
func NewMockAuthorizer(token string, valid bool) *MockAuthorizer {
	return &MockAuthorizer{
		Token:  token,
		Valid:  valid,
		Logger: NewMockLogger(),
	}
}

// GetAuthorizationHeader returns the authorization header
func (a *MockAuthorizer) GetAuthorizationHeader(ctx context.Context) (string, error) {
	if ctx != nil {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			// continue
		}
	}

	if !a.Valid {
		return "", fmt.Errorf("token is not valid")
	}

	return "Bearer " + a.Token, nil
}

// IsValid returns whether the authorizer is valid
func (a *MockAuthorizer) IsValid() bool {
	return a.Valid
}

// GetToken returns the token
func (a *MockAuthorizer) GetToken() string {
	return a.Token
}

// MockTokenManager is a simple token manager implementation for testing
type MockTokenManager struct {
	Token     string
	Valid     bool
	Logger    interfaces.Logger
	RefreshOK bool
	RevokeOK  bool
}

// NewMockTokenManager creates a new mock token manager
func NewMockTokenManager(token string, valid bool) *MockTokenManager {
	return &MockTokenManager{
		Token:     token,
		Valid:     valid,
		Logger:    NewMockLogger(),
		RefreshOK: true,
		RevokeOK:  true,
	}
}

// GetToken returns the token
func (m *MockTokenManager) GetToken(ctx context.Context) (string, error) {
	if ctx != nil {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			// continue
		}
	}

	if !m.Valid {
		return "", fmt.Errorf("token is not valid")
	}

	return m.Token, nil
}

// RefreshToken refreshes the token
func (m *MockTokenManager) RefreshToken(ctx context.Context) error {
	if ctx != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// continue
		}
	}

	if !m.RefreshOK {
		return fmt.Errorf("refresh failed")
	}

	m.Valid = true
	return nil
}

// RevokeToken revokes the token
func (m *MockTokenManager) RevokeToken(ctx context.Context) error {
	if ctx != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// continue
		}
	}

	if !m.RevokeOK {
		return fmt.Errorf("revoke failed")
	}

	m.Valid = false
	return nil
}

// IsValid returns whether the token is valid
func (m *MockTokenManager) IsValid() bool {
	return m.Valid
}

// MockConnectionPool is a simple connection pool implementation for testing
type MockConnectionPool struct {
	Client    *http.Client
	Timeout   time.Duration
	Logger    interfaces.Logger
	Transport *http.Transport
}

// NewMockConnectionPool creates a new mock connection pool
func NewMockConnectionPool() *MockConnectionPool {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     30 * time.Second,
	}

	return &MockConnectionPool{
		Client: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
		Timeout:   30 * time.Second,
		Logger:    NewMockLogger(),
		Transport: transport,
	}
}

// GetClient returns the HTTP client
func (p *MockConnectionPool) GetClient() *http.Client {
	return p.Client
}

// SetTimeout sets the timeout
func (p *MockConnectionPool) SetTimeout(timeout time.Duration) {
	p.Timeout = timeout
	p.Client.Timeout = timeout
}

// CloseIdleConnections closes all idle connections
func (p *MockConnectionPool) CloseIdleConnections() {
	p.Client.Transport.(*http.Transport).CloseIdleConnections()
}

// GetTransport returns the transport
func (p *MockConnectionPool) GetTransport() *http.Transport {
	return p.Transport
}

// MockConnectionPoolConfig is a simple connection pool config implementation for testing
type MockConnectionPoolConfig struct {
	MaxIdleConnsPerHostVal int
	MaxIdleConnsVal        int
	MaxConnsPerHostVal     int
	IdleConnTimeoutVal     time.Duration
}

// NewMockConnectionPoolConfig creates a new mock connection pool config
func NewMockConnectionPoolConfig() *MockConnectionPoolConfig {
	return &MockConnectionPoolConfig{
		MaxIdleConnsPerHostVal: 10,
		MaxIdleConnsVal:        100,
		MaxConnsPerHostVal:     100,
		IdleConnTimeoutVal:     30 * time.Second,
	}
}

// GetMaxIdleConnsPerHost returns the max idle connections per host
func (c *MockConnectionPoolConfig) GetMaxIdleConnsPerHost() int {
	return c.MaxIdleConnsPerHostVal
}

// GetMaxIdleConns returns the max idle connections
func (c *MockConnectionPoolConfig) GetMaxIdleConns() int {
	return c.MaxIdleConnsVal
}

// GetMaxConnsPerHost returns the max connections per host
func (c *MockConnectionPoolConfig) GetMaxConnsPerHost() int {
	return c.MaxConnsPerHostVal
}

// GetIdleConnTimeout returns the idle connection timeout
func (c *MockConnectionPoolConfig) GetIdleConnTimeout() time.Duration {
	return c.IdleConnTimeoutVal
}

// MockConnectionPoolManager is a simple connection pool manager implementation for testing
type MockConnectionPoolManager struct {
	Pools  map[string]interfaces.ConnectionPool
	Logger interfaces.Logger
}

// NewMockConnectionPoolManager creates a new mock connection pool manager
func NewMockConnectionPoolManager() *MockConnectionPoolManager {
	return &MockConnectionPoolManager{
		Pools:  make(map[string]interfaces.ConnectionPool),
		Logger: NewMockLogger(),
	}
}

// GetPool returns the connection pool for the given service
func (m *MockConnectionPoolManager) GetPool(serviceName string, config interfaces.ConnectionPoolConfig) interfaces.ConnectionPool {
	if pool, ok := m.Pools[serviceName]; ok {
		return pool
	}

	pool := NewMockConnectionPool()
	m.Pools[serviceName] = pool
	return pool
}

// CloseAllIdleConnections closes all idle connections across all pools
func (m *MockConnectionPoolManager) CloseAllIdleConnections() {
	for _, pool := range m.Pools {
		pool.CloseIdleConnections()
	}
}

// GetAllStats returns stats for all pools
func (m *MockConnectionPoolManager) GetAllStats() map[string]interface{} {
	stats := make(map[string]interface{})
	for name := range m.Pools {
		stats[name] = map[string]interface{}{
			"active": 0,
			"idle":   0,
		}
	}
	return stats
}

// MockPooledHTTPClient is a simple pooled HTTP client implementation for testing
type MockPooledHTTPClient struct {
	Pool interfaces.ConnectionPool
}

// NewMockPooledHTTPClient creates a new mock pooled HTTP client
func NewMockPooledHTTPClient() *MockPooledHTTPClient {
	return &MockPooledHTTPClient{
		Pool: NewMockConnectionPool(),
	}
}

// GetPool returns the connection pool
func (c *MockPooledHTTPClient) GetPool() interfaces.ConnectionPool {
	return c.Pool
}

// Do performs an HTTP request
func (c *MockPooledHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{"status": "ok"}`)),
		Header:     make(http.Header),
	}, nil
}

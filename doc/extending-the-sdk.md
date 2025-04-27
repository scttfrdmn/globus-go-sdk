<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Extending the Globus Go SDK

This guide explains how to extend the Globus Go SDK with new features, services, and integrations.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Adding New Services](#adding-new-services)
- [Customizing Authentication](#customizing-authentication)
- [Implementing Custom Token Storage](#implementing-custom-token-storage)
- [Advanced Error Handling](#advanced-error-handling)
- [Creating Middleware](#creating-middleware)
- [Testing Extensions](#testing-extensions)

## Architecture Overview

The Globus Go SDK follows a modular architecture:

```
pkg/
├── core/           # Core functionality
│   ├── auth/       # Authentication components
│   ├── config/     # Configuration
│   ├── transport/  # HTTP transport
│   └── client/     # Base client
├── services/       # Service implementations
│   ├── auth/       # Auth service
│   ├── transfer/   # Transfer service
│   ├── groups/     # Groups service
│   └── ...         # Other services
└── utils/          # Utility functions
```

Each service follows a similar structure:

```
services/example/
├── client.go       # Client implementation
├── client_test.go  # Client tests
├── models.go       # Data models
├── operations.go   # API operations
└── errors.go       # Service-specific errors
```

## Adding New Services

To add a new service to the SDK:

### 1. Create the Service Package

```go
// pkg/services/example/models.go
package example

// Item represents an example item
type Item struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description,omitempty"`
    Created     time.Time `json:"created"`
}

// ListItemsOptions contains parameters for the ListItems method
type ListItemsOptions struct {
    Limit  int    `json:"limit,omitempty"`
    Marker string `json:"marker,omitempty"`
}

// ToQueryParams converts options to query parameters
func (o *ListItemsOptions) ToQueryParams() map[string]string {
    params := make(map[string]string)
    if o.Limit > 0 {
        params["limit"] = strconv.Itoa(o.Limit)
    }
    if o.Marker != "" {
        params["marker"] = o.Marker
    }
    return params
}
```

### 2. Implement the Client

```go
// pkg/services/example/client.go
package example

import (
    "context"
    "net/http"
    "net/url"

    "github.com/scttfrdmn/globus-go-sdk/pkg/core/auth"
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/client"
)

const (
    // DefaultBaseURL is the default base URL for the Example service
    DefaultBaseURL = "https://example.globus.org/v1/"
)

// Client provides access to the Example API
type Client struct {
    baseURL    *url.URL
    httpClient *http.Client
    authorizer auth.Authorizer
}

// NewClient creates a new Example client
func NewClient(authorizer auth.Authorizer) *Client {
    baseURL, _ := url.Parse(DefaultBaseURL)
    
    return &Client{
        baseURL:    baseURL,
        httpClient: &http.Client{},
        authorizer: authorizer,
    }
}

// WithBaseURL sets the base URL for the client
func (c *Client) WithBaseURL(baseURL string) (*Client, error) {
    u, err := url.Parse(baseURL)
    if err != nil {
        return nil, err
    }
    
    c.baseURL = u
    return c, nil
}

// WithHTTPClient sets the HTTP client for the client
func (c *Client) WithHTTPClient(httpClient *http.Client) *Client {
    c.httpClient = httpClient
    return c
}
```

### 3. Implement API Operations

```go
// pkg/services/example/operations.go
package example

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "path"

    "github.com/scttfrdmn/globus-go-sdk/pkg/core/client"
)

// ListItems retrieves a list of items
func (c *Client) ListItems(ctx context.Context, options *ListItemsOptions) ([]Item, error) {
    if options == nil {
        options = &ListItemsOptions{}
    }
    
    // Build request
    u := *c.baseURL
    u.Path = path.Join(u.Path, "items")
    
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
    if err != nil {
        return nil, fmt.Errorf("error creating request: %w", err)
    }
    
    // Add query parameters
    q := req.URL.Query()
    for key, value := range options.ToQueryParams() {
        q.Add(key, value)
    }
    req.URL.RawQuery = q.Encode()
    
    // Add authorization
    if err := c.authorizer.AddToRequest(req); err != nil {
        return nil, fmt.Errorf("error adding authorization: %w", err)
    }
    
    // Execute request
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error executing request: %w", err)
    }
    defer resp.Body.Close()
    
    // Check for errors
    if resp.StatusCode != http.StatusOK {
        var apiErr client.APIError
        if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
            return nil, fmt.Errorf("error parsing error response: %w", err)
        }
        return nil, &apiErr
    }
    
    // Parse response
    var items []Item
    if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
        return nil, fmt.Errorf("error parsing response: %w", err)
    }
    
    return items, nil
}

// GetItem retrieves a single item by ID
func (c *Client) GetItem(ctx context.Context, id string) (*Item, error) {
    // Implementation similar to ListItems
    // ...
}

// CreateItem creates a new item
func (c *Client) CreateItem(ctx context.Context, item *Item) (*Item, error) {
    // Implementation for POST request
    // ...
}

// UpdateItem updates an existing item
func (c *Client) UpdateItem(ctx context.Context, id string, item *Item) (*Item, error) {
    // Implementation for PUT request
    // ...
}

// DeleteItem deletes an item
func (c *Client) DeleteItem(ctx context.Context, id string) error {
    // Implementation for DELETE request
    // ...
}
```

### 4. Add Service-Specific Errors

```go
// pkg/services/example/errors.go
package example

import (
    "errors"
    "strings"

    "github.com/scttfrdmn/globus-go-sdk/pkg/core/client"
)

// Common error codes
const (
    ErrorCodeItemNotFound      = "ItemNotFound"
    ErrorCodeInvalidParameters = "InvalidParameters"
    ErrorCodePermissionDenied  = "PermissionDenied"
)

// IsItemNotFoundError checks if the error is an item not found error
func IsItemNotFoundError(err error) bool {
    var apiErr *client.APIError
    if errors.As(err, &apiErr) {
        return apiErr.Code == ErrorCodeItemNotFound
    }
    return false
}

// IsPermissionDeniedError checks if the error is a permission denied error
func IsPermissionDeniedError(err error) bool {
    var apiErr *client.APIError
    if errors.As(err, &apiErr) {
        return apiErr.Code == ErrorCodePermissionDenied
    }
    return false
}

// IsInvalidParametersError checks if the error is an invalid parameters error
func IsInvalidParametersError(err error) bool {
    var apiErr *client.APIError
    if errors.As(err, &apiErr) {
        return apiErr.Code == ErrorCodeInvalidParameters
    }
    return false
}
```

### 5. Write Tests

```go
// pkg/services/example/client_test.go
package example

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/scttfrdmn/globus-go-sdk/pkg/core/auth"
)

func TestListItems(t *testing.T) {
    // Create a test server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Check request method
        if r.Method != http.MethodGet {
            t.Errorf("Expected GET request, got %s", r.Method)
        }
        
        // Check authorization header
        if r.Header.Get("Authorization") != "Bearer test-token" {
            t.Errorf("Expected Authorization header 'Bearer test-token', got %s", r.Header.Get("Authorization"))
        }
        
        // Check query parameters
        if r.URL.Query().Get("limit") != "10" {
            t.Errorf("Expected limit=10, got %s", r.URL.Query().Get("limit"))
        }
        
        // Return mock response
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`[
            {"id": "item1", "name": "Item 1", "created": "2023-01-01T00:00:00Z"},
            {"id": "item2", "name": "Item 2", "created": "2023-01-02T00:00:00Z"}
        ]`))
    }))
    defer server.Close()
    
    // Create client with test server URL
    authorizer := auth.NewBearerTokenAuthorizer("test-token")
    client := NewClient(authorizer)
    client, _ = client.WithBaseURL(server.URL)
    
    // Call the method
    items, err := client.ListItems(context.Background(), &ListItemsOptions{
        Limit: 10,
    })
    
    // Check for errors
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
    
    // Check response
    if len(items) != 2 {
        t.Errorf("Expected 2 items, got %d", len(items))
    }
    if items[0].ID != "item1" {
        t.Errorf("Expected ID 'item1', got %s", items[0].ID)
    }
    if items[0].Name != "Item 1" {
        t.Errorf("Expected Name 'Item 1', got %s", items[0].Name)
    }
}
```

## Customizing Authentication

You can create custom authorizers by implementing the `auth.Authorizer` interface:

```go
// pkg/core/auth/custom_authorizer.go
package auth

import (
    "net/http"
)

// CustomAuthorizer is a custom implementation of the Authorizer interface
type CustomAuthorizer struct {
    // Your fields here
}

// NewCustomAuthorizer creates a new custom authorizer
func NewCustomAuthorizer() *CustomAuthorizer {
    return &CustomAuthorizer{}
}

// AddToRequest adds authorization to the request
func (a *CustomAuthorizer) AddToRequest(req *http.Request) error {
    // Add your custom authorization logic here
    // For example:
    req.Header.Set("X-Custom-Auth", "your-custom-auth-value")
    return nil
}
```

## Implementing Custom Token Storage

You can create custom token storage implementations by implementing the `auth.TokenStorage` interface:

```go
// pkg/core/auth/custom_storage.go
package auth

import (
    "context"
)

// DatabaseTokenStorage implements TokenStorage using a database
type DatabaseTokenStorage struct {
    // Your fields here (e.g., database connection)
}

// NewDatabaseTokenStorage creates a new database token storage
func NewDatabaseTokenStorage(connString string) (*DatabaseTokenStorage, error) {
    // Initialize your database connection
    return &DatabaseTokenStorage{
        // Initialize fields
    }, nil
}

// StoreToken stores a token in the database
func (s *DatabaseTokenStorage) StoreToken(ctx context.Context, key string, token TokenInfo) error {
    // Store the token in your database
    return nil
}

// GetToken retrieves a token from the database
func (s *DatabaseTokenStorage) GetToken(ctx context.Context, key string) (TokenInfo, error) {
    // Retrieve the token from your database
    return TokenInfo{}, nil
}

// DeleteToken deletes a token from the database
func (s *DatabaseTokenStorage) DeleteToken(ctx context.Context, key string) error {
    // Delete the token from your database
    return nil
}

// ListTokens lists all token keys in the database
func (s *DatabaseTokenStorage) ListTokens(ctx context.Context) ([]string, error) {
    // List all token keys from your database
    return []string{}, nil
}
```

## Advanced Error Handling

You can create custom error types for your extensions:

```go
// pkg/services/example/errors.go
package example

import (
    "fmt"
    "net/http"
)

// ValidationError represents a validation error
type ValidationError struct {
    Field   string
    Message string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error for field %s: %s", e.Field, e.Message)
}

// RetryableError represents an error that can be retried
type RetryableError struct {
    Cause      error
    RetryAfter int
}

// Error implements the error interface
func (e *RetryableError) Error() string {
    return fmt.Sprintf("retryable error: %v (retry after %d seconds)", e.Cause, e.RetryAfter)
}

// Unwrap returns the underlying error
func (e *RetryableError) Unwrap() error {
    return e.Cause
}

// IsRetryableError checks if an error is a retryable error
func IsRetryableError(err error) bool {
    _, ok := err.(*RetryableError)
    return ok
}
```

## Creating Middleware

You can create custom HTTP middleware for logging, metrics, etc.:

```go
// pkg/core/transport/middleware.go
package transport

import (
    "net/http"
    "time"
)

// Middleware defines an HTTP middleware function
type Middleware func(http.RoundTripper) http.RoundTripper

// LoggingTransport logs HTTP requests and responses
type LoggingTransport struct {
    Next   http.RoundTripper
    Logger Logger
}

// RoundTrip implements the http.RoundTripper interface
func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    // Log the request
    t.Logger.Infof("Request: %s %s", req.Method, req.URL.String())
    
    // Call the next transport
    start := time.Now()
    resp, err := t.Next.RoundTrip(req)
    duration := time.Since(start)
    
    // Log the response
    if err != nil {
        t.Logger.Errorf("Error: %v", err)
    } else {
        t.Logger.Infof("Response: %d %s (%s)", resp.StatusCode, resp.Status, duration)
    }
    
    return resp, err
}

// WithLogging adds logging middleware to an HTTP client
func WithLogging(logger Logger) Middleware {
    return func(next http.RoundTripper) http.RoundTripper {
        return &LoggingTransport{
            Next:   next,
            Logger: logger,
        }
    }
}

// ApplyMiddleware applies middleware to an HTTP client
func ApplyMiddleware(client *http.Client, middleware ...Middleware) *http.Client {
    transport := client.Transport
    if transport == nil {
        transport = http.DefaultTransport
    }
    
    for i := len(middleware) - 1; i >= 0; i-- {
        transport = middleware[i](transport)
    }
    
    result := *client
    result.Transport = transport
    return &result
}
```

Usage:

```go
// Create a client with middleware
httpClient := &http.Client{}
httpClient = transport.ApplyMiddleware(httpClient, 
    transport.WithLogging(logger),
    transport.WithRetry(3, 1*time.Second),
)

// Create a service client with the custom HTTP client
client := example.NewClient(authorizer).WithHTTPClient(httpClient)
```

## Testing Extensions

### Unit Testing

```go
// pkg/services/example/custom_test.go
package example

import (
    "context"
    "errors"
    "testing"
)

func TestCustomError(t *testing.T) {
    // Create a test error
    err := &ValidationError{
        Field:   "name",
        Message: "cannot be empty",
    }
    
    // Check the error message
    if err.Error() != "validation error for field name: cannot be empty" {
        t.Errorf("Unexpected error message: %s", err.Error())
    }
}

func TestRetryableError(t *testing.T) {
    // Create a test error
    cause := errors.New("connection reset")
    err := &RetryableError{
        Cause:      cause,
        RetryAfter: 5,
    }
    
    // Check if the error is retryable
    if !IsRetryableError(err) {
        t.Error("Expected error to be retryable")
    }
    
    // Check if the cause is correctly unwrapped
    if !errors.Is(err, cause) {
        t.Error("Expected error to unwrap to cause")
    }
}
```

### Integration Testing

```go
// pkg/services/example/integration_test.go
package example

import (
    "context"
    "os"
    "testing"

    "github.com/scttfrdmn/globus-go-sdk/pkg/core/auth"
)

// skipIfNoToken skips the test if there's no token available
func skipIfNoToken(t *testing.T) string {
    token := os.Getenv("EXAMPLE_TEST_TOKEN")
    if token == "" {
        t.Skip("Skipping integration test (EXAMPLE_TEST_TOKEN not set)")
    }
    return token
}

func TestIntegrationListItems(t *testing.T) {
    token := skipIfNoToken(t)
    
    // Create client with real token
    authorizer := auth.NewBearerTokenAuthorizer(token)
    client := NewClient(authorizer)
    
    // Call the real API
    items, err := client.ListItems(context.Background(), nil)
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    
    // Validate response
    if len(items) == 0 {
        t.Log("No items returned, but no error occurred")
    } else {
        t.Logf("Found %d items", len(items))
    }
}
```

## Best Practices for Extensions

1. **Follow Existing Patterns**: Maintain consistency with the rest of the SDK.

2. **Use Context**: All operations should accept a context.Context as the first parameter.

3. **Provide Options**: Use functional options or option structs for extensibility.

4. **Robust Error Handling**: Return structured errors with adequate information.

5. **Strong Typing**: Use strong typing for API models and parameters.

6. **Documentation**: Document all exported functions, types, and fields.

7. **Testing**: Write both unit and integration tests.

8. **Versioning**: Follow semantic versioning for your extensions.

9. **Zero Dependencies**: Minimize external dependencies.

10. **Backward Compatibility**: Ensure changes don't break existing code.
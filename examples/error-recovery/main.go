// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/core/ratelimit"
)

// Configuration options for the example
type Config struct {
	// Simulation options
	UseRealServices bool
	SimulateError   string
	FailureRate     float64

	// Circuit breaker settings
	FailureThreshold int
	ResetTimeout     time.Duration

	// Retry settings
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	BackoffFactor  float64
	Jitter         float64

	// Resource management
	PoolSize        int
	HealthCheckFreq time.Duration

	// Authentication
	ClientID     string
	ClientSecret string
	RefreshToken string
}

// Application state
type App struct {
	Config      Config
	CircuitOpen bool
	FailCount   int
	LastFailure time.Time
	RetryStats  map[string]int
	Mutex       sync.RWMutex
	SDK         *pkg.SDKConfig
}

// Initialize the application with default config
func NewApp() *App {
	return &App{
		Config: Config{
			UseRealServices:  false,
			SimulateError:    "",
			FailureRate:      0.2,
			FailureThreshold: 5,
			ResetTimeout:     time.Minute,
			MaxRetries:       10,
			InitialBackoff:   100 * time.Millisecond,
			MaxBackoff:       30 * time.Second,
			BackoffFactor:    2.0,
			Jitter:           0.1,
			PoolSize:         10,
			HealthCheckFreq:  30 * time.Second,
			ClientID:         os.Getenv("GLOBUS_CLIENT_ID"),
			ClientSecret:     os.Getenv("GLOBUS_CLIENT_SECRET"),
			RefreshToken:     os.Getenv("GLOBUS_REFRESH_TOKEN"),
		},
		CircuitOpen: false,
		FailCount:   0,
		RetryStats:  make(map[string]int),
	}
}

// Parse command line flags to update config
func (app *App) ParseFlags() {
	flag.BoolVar(&app.Config.UseRealServices, "use-real-services", app.Config.UseRealServices, "Use real Globus services instead of simulation")
	flag.StringVar(&app.Config.SimulateError, "simulate", app.Config.SimulateError, "Simulate a specific error condition (network-partition, auth-failure, rate-limit, timeout)")
	flag.Float64Var(&app.Config.FailureRate, "failure-rate", app.Config.FailureRate, "Probability of simulated failures (0.0-1.0)")
	flag.Parse()
}

// Initialize the Globus SDK
func (app *App) InitSDK() error {
	// Create SDK configuration
	config := pkg.NewConfig().
		WithClientID(app.Config.ClientID).
		WithClientSecret(app.Config.ClientSecret)

	if app.Config.UseRealServices {
		// Use environment settings when available
		envConfig := pkg.NewConfigFromEnvironment().
			WithClientID(app.Config.ClientID).
			WithClientSecret(app.Config.ClientSecret)
		app.SDK = envConfig
	} else {
		// Use basic configuration
		app.SDK = config
	}

	return nil
}

// Run demonstrates error recovery patterns
func (app *App) Run() error {
	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Show config
	app.logConfig()

	// Demonstrate patterns
	err := app.demonstrateCircuitBreaker(ctx)
	if err != nil {
		log.Printf("Circuit breaker demo failed: %v", err)
	}

	err = app.demonstrateRetryWithBackoff(ctx)
	if err != nil {
		log.Printf("Retry with backoff demo failed: %v", err)
	}

	err = app.demonstrateGracefulDegradation(ctx)
	if err != nil {
		log.Printf("Graceful degradation demo failed: %v", err)
	}

	err = app.demonstrateAuthRecovery(ctx)
	if err != nil {
		log.Printf("Auth recovery demo failed: %v", err)
	}

	err = app.demonstrateResourceManagement(ctx)
	if err != nil {
		log.Printf("Resource management demo failed: %v", err)
	}

	return nil
}

// Log the current configuration
func (app *App) logConfig() {
	log.Println("=== Error Recovery Example Configuration ===")
	log.Printf("Simulation: UseRealServices=%v, SimulateError=%q, FailureRate=%.2f\n",
		app.Config.UseRealServices, app.Config.SimulateError, app.Config.FailureRate)
	log.Printf("Circuit Breaker: FailureThreshold=%d, ResetTimeout=%v\n",
		app.Config.FailureThreshold, app.Config.ResetTimeout)
	log.Printf("Retry: MaxRetries=%d, InitialBackoff=%v, MaxBackoff=%v, Factor=%.1f, Jitter=%.1f\n",
		app.Config.MaxRetries, app.Config.InitialBackoff, app.Config.MaxBackoff,
		app.Config.BackoffFactor, app.Config.Jitter)
	log.Printf("Resources: PoolSize=%d, HealthCheckFreq=%v\n",
		app.Config.PoolSize, app.Config.HealthCheckFreq)
	log.Println("============================================")
}

// Simulate an operation that may fail
func (app *App) simulateOperation(ctx context.Context, name string) error {
	// If using real services, perform actual operation
	if app.Config.UseRealServices {
		return app.performRealOperation(ctx, name)
	}

	// Check if we should simulate a specific error
	if app.Config.SimulateError != "" {
		return app.simulateSpecificError(name)
	}

	// Random failure based on failure rate
	if rand.Float64() < app.Config.FailureRate {
		return fmt.Errorf("simulated failure in %s operation", name)
	}

	// Simulate operation latency
	time.Sleep(time.Duration(50+rand.Intn(200)) * time.Millisecond)
	return nil
}

// Perform a real operation against Globus services
func (app *App) performRealOperation(ctx context.Context, name string) error {
	switch name {
	case "auth-user-info":
		authClient, err := app.SDK.NewAuthClient()
		if err != nil {
			return err
		}
		_, err = authClient.GetUserInfo(ctx, "dummy-token")
		return err
	case "auth-token-introspect":
		authClient, err := app.SDK.NewAuthClient()
		if err != nil {
			return err
		}
		_, err = authClient.IntrospectToken(ctx, "dummy-token")
		return err
	case "transfer-list-endpoints":
		transferClient, err := app.SDK.NewTransferClient("dummy-token")
		if err != nil {
			return err
		}
		_, err = transferClient.ListEndpoints(ctx, nil)
		return err
	default:
		return fmt.Errorf("unknown operation: %s", name)
	}
}

// Simulate specific error conditions
func (app *App) simulateSpecificError(name string) error {
	switch app.Config.SimulateError {
	case "network-partition":
		return fmt.Errorf("network unavailable: connection refused")
	case "auth-failure":
		return fmt.Errorf("UNAUTHORIZED: The request lacks valid authentication credentials")
	case "rate-limit":
		return fmt.Errorf("TOO_MANY_REQUESTS: Rate limit exceeded")
	case "timeout":
		return context.DeadlineExceeded
	default:
		return fmt.Errorf("unknown simulated error: %s", app.Config.SimulateError)
	}
}

// Demonstrate the circuit breaker pattern
func (app *App) demonstrateCircuitBreaker(ctx context.Context) error {
	log.Println("\n=== Circuit Breaker Pattern Demo ===")

	// Create a circuit breaker
	cb := ratelimit.NewCircuitBreaker(&ratelimit.CircuitBreakerOptions{
		Threshold: app.Config.FailureThreshold,
		Timeout:   app.Config.ResetTimeout,
	})

	// Run operations through the circuit breaker
	for i := 0; i < 20; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Check if circuit is open
			if !cb.AllowRequest() {
				log.Printf("Circuit is OPEN, skipping request %d", i)
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Attempt operation
			log.Printf("Attempt %d: Executing operation through circuit breaker", i)
			err := app.simulateOperation(ctx, "auth-user-info")

			if err != nil {
				log.Printf("Attempt %d: FAILED - %v", i, err)
				cb.RecordResult(err)
			} else {
				log.Printf("Attempt %d: SUCCESS", i)
				cb.RecordResult(nil)
			}

			time.Sleep(200 * time.Millisecond)
		}
	}

	return nil
}

// Demonstrate retry with exponential backoff
func (app *App) demonstrateRetryWithBackoff(ctx context.Context) error {
	log.Println("\n=== Retry with Exponential Backoff Demo ===")

	// Create backoff strategy
	backoff := ratelimit.NewExponentialBackoff(
		app.Config.InitialBackoff,
		app.Config.MaxBackoff,
		app.Config.BackoffFactor,
		app.Config.MaxRetries,
	)

	// Add jitter if configured
	backoff.Jitter = true
	backoff.JitterFactor = app.Config.Jitter

	// Execute operation with retries
	var lastErr error
	for attempt := 0; attempt <= app.Config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if attempt > 0 {
				delay := backoff.NextBackoff(attempt)
				log.Printf("Retry %d/%d: Waiting %v before retry",
					attempt, app.Config.MaxRetries, delay)
				time.Sleep(delay)
			}

			log.Printf("Attempt %d/%d: Executing operation with retry logic",
				attempt, app.Config.MaxRetries)
			err := app.simulateOperation(ctx, "transfer-list-endpoints")

			if err == nil {
				log.Printf("Attempt %d/%d: SUCCESS", attempt, app.Config.MaxRetries)
				return nil
			}

			log.Printf("Attempt %d/%d: FAILED - %v", attempt, app.Config.MaxRetries, err)
			lastErr = err

			// Check if error is retriable
			if !isRetriableError(err) {
				log.Printf("Non-retriable error, giving up: %v", err)
				return err
			}
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// Demonstrate graceful degradation
func (app *App) demonstrateGracefulDegradation(ctx context.Context) error {
	log.Println("\n=== Graceful Degradation Demo ===")

	// Define features with priority and fallback options
	features := []struct {
		name      string
		priority  int // 1=critical, 2=important, 3=optional
		operation string
		fallback  func() (string, error)
	}{
		{
			name:      "user-profile",
			priority:  1,
			operation: "auth-user-info",
			fallback: func() (string, error) {
				return "cached-user-profile", nil
			},
		},
		{
			name:      "endpoint-list",
			priority:  2,
			operation: "transfer-list-endpoints",
			fallback: func() (string, error) {
				return "cached-endpoint-list", nil
			},
		},
		{
			name:      "token-validation",
			priority:  3,
			operation: "auth-token-introspect",
			fallback: func() (string, error) {
				return "", fmt.Errorf("no fallback available")
			},
		},
	}

	// Try each feature with degradation if needed
	for _, feature := range features {
		log.Printf("Attempting feature: %s (priority: %d)",
			feature.name, feature.priority)

		err := app.simulateOperation(ctx, feature.operation)
		if err == nil {
			log.Printf("Feature %s: SUCCESS using primary implementation",
				feature.name)
			continue
		}

		log.Printf("Feature %s: Primary implementation FAILED - %v",
			feature.name, err)

		// Try fallback for non-critical features
		if feature.priority > 1 {
			result, fbErr := feature.fallback()
			if fbErr == nil {
				log.Printf("Feature %s: DEGRADED - using fallback: %s",
					feature.name, result)
			} else {
				log.Printf("Feature %s: UNAVAILABLE - fallback also failed: %v",
					feature.name, fbErr)
			}
		} else {
			// Critical feature failure
			log.Printf("Feature %s: CRITICAL FAILURE - operation cannot continue",
				feature.name)
			return err
		}
	}

	log.Println("Application running in degraded mode with fallbacks")
	return nil
}

// Demonstrate authentication failure recovery
func (app *App) demonstrateAuthRecovery(ctx context.Context) error {
	log.Println("\n=== Authentication Failure Recovery Demo ===")

	// Simulate token expiration and refresh
	log.Println("Simulating expired token...")
	originalToken := "expired-token"
	refreshedToken := "new-token"

	// Attempt operation with expired token
	log.Printf("Attempting operation with token: %s", originalToken)
	err := app.simulateOperation(ctx, "auth-user-info")

	if err == nil {
		log.Println("Unexpected success with expired token")
		return nil
	}

	// Check if this is an auth error
	if isAuthError(err) {
		log.Printf("Authentication failed: %v", err)
		log.Println("Attempting token refresh...")

		// Simulate token refresh
		time.Sleep(300 * time.Millisecond)
		log.Printf("Token refreshed successfully: %s", refreshedToken)

		// Retry operation with new token
		log.Printf("Retrying operation with new token: %s", refreshedToken)
		retryErr := app.simulateOperation(ctx, "auth-user-info")

		if retryErr == nil {
			log.Println("Operation succeeded after token refresh")
			return nil
		} else {
			log.Printf("Operation failed even after token refresh: %v", retryErr)
			return retryErr
		}
	}

	return err
}

// Demonstrate resource management during failures
func (app *App) demonstrateResourceManagement(ctx context.Context) error {
	log.Println("\n=== Resource Management During Failures Demo ===")

	// Simulate connection pool status
	type connectionStatus struct {
		id       int
		active   bool
		lastUsed time.Time
		error    error
	}

	// Initialize connection pool
	pool := make([]connectionStatus, app.Config.PoolSize)
	for i := range pool {
		pool[i] = connectionStatus{
			id:       i,
			active:   true,
			lastUsed: time.Now().Add(-time.Duration(rand.Intn(300)) * time.Second),
			error:    nil,
		}
	}

	// Display initial pool status
	log.Println("Initial connection pool status:")
	for _, conn := range pool {
		log.Printf("Connection ID: %d, Active: %v, Last Used: %v",
			conn.id, conn.active, time.Since(conn.lastUsed).Round(time.Second))
	}

	// Simulate some failures
	failCount := int(float64(app.Config.PoolSize) * app.Config.FailureRate)
	log.Printf("Simulating %d connection failures...", failCount)

	for i := 0; i < failCount; i++ {
		idx := rand.Intn(app.Config.PoolSize)
		pool[idx].active = false
		pool[idx].error = fmt.Errorf("connection error: broken pipe")
	}

	// Display pool after failures
	log.Println("Connection pool status after failures:")
	for _, conn := range pool {
		if conn.active {
			log.Printf("Connection ID: %d, Active: %v, Last Used: %v",
				conn.id, conn.active, time.Since(conn.lastUsed).Round(time.Second))
		} else {
			log.Printf("Connection ID: %d, Active: %v, Error: %v",
				conn.id, conn.active, conn.error)
		}
	}

	// Simulate pool recovery
	log.Println("Performing connection pool recovery...")

	// 1. Close failed connections
	for i := range pool {
		if !pool[i].active {
			log.Printf("Closing failed connection ID: %d", pool[i].id)
			// In real code, would close the connection here
			time.Sleep(50 * time.Millisecond)
		}
	}

	// 2. Refresh idle connections
	idleTimeout := 2 * time.Minute
	for i := range pool {
		if pool[i].active && time.Since(pool[i].lastUsed) > idleTimeout {
			log.Printf("Refreshing idle connection ID: %d (idle for %v)",
				pool[i].id, time.Since(pool[i].lastUsed).Round(time.Second))
			// In real code, would refresh the connection here
			pool[i].lastUsed = time.Now()
			time.Sleep(50 * time.Millisecond)
		}
	}

	// 3. Create new connections to replace failed ones
	for i := range pool {
		if !pool[i].active {
			log.Printf("Creating new connection to replace ID: %d", pool[i].id)
			// In real code, would create a new connection here
			pool[i].active = true
			pool[i].error = nil
			pool[i].lastUsed = time.Now()
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Display final pool status
	log.Println("Final connection pool status after recovery:")
	activeCount := 0
	for _, conn := range pool {
		log.Printf("Connection ID: %d, Active: %v, Last Used: %v",
			conn.id, conn.active, time.Since(conn.lastUsed).Round(time.Second))
		if conn.active {
			activeCount++
		}
	}

	log.Printf("Pool recovery complete: %d/%d connections active",
		activeCount, app.Config.PoolSize)

	return nil
}

// Helper function to identify retriable errors
func isRetriableError(err error) bool {
	// Network errors are generally retriable
	if err != nil && (err.Error() == "network unavailable: connection refused" ||
		err.Error() == "i/o timeout" ||
		err == context.DeadlineExceeded) {
		return true
	}

	// Check for rate limit errors in error message
	if err != nil && (strings.Contains(err.Error(), "TOO_MANY_REQUESTS") ||
		strings.Contains(err.Error(), "RESOURCE_TEMPORARILY_UNAVAILABLE")) {
		return true
	}

	return false
}

// Helper function to identify authentication errors
func isAuthError(err error) bool {
	if err != nil && (strings.Contains(err.Error(), "UNAUTHORIZED") ||
		strings.Contains(err.Error(), "TOKEN_EXPIRED")) {
		return true
	}
	return false
}

func main() {
	// Initialize random with Go 1.20+ approach (no explicit seed needed)
	// The math/rand package automatically seeds the global source since Go 1.20

	// Create and run the app
	app := NewApp()
	app.ParseFlags()

	err := app.InitSDK()
	if err != nil {
		log.Fatalf("Failed to initialize SDK: %v", err)
	}

	err = app.Run()
	if err != nil {
		log.Fatalf("Application failed: %v", err)
	}

	log.Println("Example completed successfully")
}

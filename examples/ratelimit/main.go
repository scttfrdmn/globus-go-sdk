// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/yourusername/globus-go-sdk/pkg/core/ratelimit"
	"github.com/yourusername/globus-go-sdk/pkg/services/auth"
	"github.com/yourusername/globus-go-sdk/pkg/services/transfer"
)

func main() {
	// Parse command line flags
	mode := flag.String("mode", "demo", "Mode to run: demo, ratelimit, backoff, circuit")
	concurrency := flag.Int("concurrency", 10, "Number of concurrent requests")
	duration := flag.Int("duration", 10, "Test duration in seconds")
	reqPerSec := flag.Float64("rate", 5.0, "Requests per second")
	accessToken := flag.String("token", "", "Globus access token (if not provided, will use auth flow)")
	
	flag.Parse()
	
	// Get access token
	tokenStr := *accessToken
	if tokenStr == "" {
		token, err := getAccessToken()
		if err != nil {
			log.Fatalf("Error getting access token: %v", err)
		}
		tokenStr = token
	}
	
	// Run the selected mode
	switch *mode {
	case "demo":
		runDemoMode(tokenStr, *concurrency, time.Duration(*duration)*time.Second, *reqPerSec)
	case "ratelimit":
		runRateLimitTest(tokenStr, *concurrency, time.Duration(*duration)*time.Second, *reqPerSec)
	case "backoff":
		runBackoffTest()
	case "circuit":
		runCircuitBreakerTest()
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		flag.Usage()
		os.Exit(1)
	}
}

// runDemoMode demonstrates all rate limiting features in one test
func runDemoMode(accessToken string, concurrency int, duration time.Duration, reqPerSec float64) {
	fmt.Println("=== Running Rate Limit and Backoff Demo ===")
	fmt.Printf("Concurrency: %d, Duration: %s, Rate: %.1f req/sec\n", 
		concurrency, duration, reqPerSec)
	
	// Create rate limiter
	options := &ratelimit.RateLimiterOptions{
		RequestsPerSecond: reqPerSec,
		BurstSize:         int(reqPerSec * 2),
		UseAdaptive:       true,
		MaxRetryCount:     3,
		MinRetryDelay:     100 * time.Millisecond,
		MaxRetryDelay:     5 * time.Second,
		UseJitter:         true,
	}
	
	limiter := ratelimit.NewTokenBucketLimiter(options)
	
	// Create circuit breaker
	cbOptions := &ratelimit.CircuitBreakerOptions{
		Threshold:         5,
		Timeout:           10 * time.Second,
		HalfOpenSuccesses: 2,
		OnStateChange: func(from, to ratelimit.CircuitBreakerState) {
			fmt.Printf("Circuit breaker state changed from %v to %v\n", from, to)
		},
	}
	
	cb := ratelimit.NewCircuitBreaker(cbOptions)
	
	// Create backoff strategy
	backoff := ratelimit.NewExponentialBackoff(
		100*time.Millisecond,
		5*time.Second,
		2.0,
		3,
	)
	
	// Set up stats tracking
	var (
		mu             sync.Mutex
		totalRequests  int
		successCount   int
		backoffCount   int
		circuitOpens   int
		rateLimitCount int
	)
	
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	
	// Create worker pool
	var wg sync.WaitGroup
	
	// Start time
	startTime := time.Now()
	
	// Create a channel for worker tasks
	tasks := make(chan int, concurrency*10)
	
	// Launch workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for task := range tasks {
				// Track request
				mu.Lock()
				totalRequests++
				reqNum := totalRequests
				mu.Unlock()
				
				// Use circuit breaker
				err := cb.Execute(ctx, func(ctx context.Context) error {
					// Wait for rate limiter
					err := limiter.Wait(ctx)
					if err != nil {
						mu.Lock()
						rateLimitCount++
						mu.Unlock()
						return err
					}
					
					// Simulate API call with retry logic
					return ratelimit.RetryWithBackoff(ctx, func(ctx context.Context) error {
						// Simulate some work
						time.Sleep(50 * time.Millisecond)
						
						// Simulate random failures (10% chance)
						if rand.Float64() < 0.1 {
							mu.Lock()
							backoffCount++
							mu.Unlock()
							return errors.New("temporary error: service unavailable")
						}
						
						// Success
						mu.Lock()
						successCount++
						mu.Unlock()
						return nil
						
					}, backoff, ratelimit.IsRetryableError)
				})
				
				if errors.Is(err, ratelimit.ErrCircuitOpen) {
					mu.Lock()
					circuitOpens++
					mu.Unlock()
				}
				
				// Print periodic status
				if reqNum%20 == 0 {
					elapsed := time.Since(startTime)
					rate := float64(reqNum) / elapsed.Seconds()
					
					mu.Lock()
					fmt.Printf("[%s] Requests: %d, Success: %d, Rate: %.1f req/sec, Retries: %d, Circuit Opens: %d\n",
						formatDuration(elapsed), reqNum, successCount, rate, backoffCount, circuitOpens)
					
					// Print rate limiter stats
					stats := limiter.GetStats()
					fmt.Printf("  Rate Limiter: Limit=%.1f, Remaining=%.1f, Throttled=%d\n",
						stats.CurrentLimit, stats.RemainingTokens, stats.TotalThrottled)
					mu.Unlock()
				}
			}
		}(i)
	}
	
	// Feed tasks to workers
	taskID := 0
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			close(tasks)
			wg.Wait()
			
			// Print final stats
			elapsed := time.Since(startTime)
			rate := float64(totalRequests) / elapsed.Seconds()
			
			fmt.Println("\n=== Final Statistics ===")
			fmt.Printf("Total Duration: %s\n", formatDuration(elapsed))
			fmt.Printf("Total Requests: %d\n", totalRequests)
			fmt.Printf("Successful Requests: %d\n", successCount)
			fmt.Printf("Average Rate: %.2f req/sec\n", rate)
			fmt.Printf("Retried Requests: %d\n", backoffCount)
			fmt.Printf("Circuit Breaker Opens: %d\n", circuitOpens)
			
			limiterStats := limiter.GetStats()
			fmt.Printf("Rate Limiter Throttled: %d\n", limiterStats.TotalThrottled)
			fmt.Printf("Total Wait Time: %s\n", limiterStats.TotalWaitTime)
			
			return
			
		case <-ticker.C:
			select {
			case tasks <- taskID:
				taskID++
			default:
				// Channel is full, skip this tick
			}
		}
	}
}

// runRateLimitTest tests the rate limiter with a real Globus API
func runRateLimitTest(accessToken string, concurrency int, duration time.Duration, reqPerSec float64) {
	fmt.Println("=== Running Rate Limit Test with Globus API ===")
	fmt.Printf("Concurrency: %d, Duration: %s, Target Rate: %.1f req/sec\n", 
		concurrency, duration, reqPerSec)
	
	// Create rate limiter
	options := &ratelimit.RateLimiterOptions{
		RequestsPerSecond: reqPerSec,
		BurstSize:         int(reqPerSec * 2),
		UseAdaptive:       true,
	}
	
	limiter := ratelimit.NewTokenBucketLimiter(options)
	
	// Track metrics
	var (
		mu            sync.Mutex
		totalRequests int
		successCount  int
		errorCount    int
		waitTime      time.Duration
	)
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	
	// Start workers
	var wg sync.WaitGroup
	startTime := time.Now()
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			// Create Globus transfer client (would normally use this for API calls)
			// Not making actual transfer calls to avoid modifying user data
			client := transfer.NewClient(accessToken)
			
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Check if we should continue making requests
					waitStart := time.Now()
					err := limiter.Wait(ctx)
					if err != nil {
						if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
							fmt.Printf("Rate limiter error: %v\n", err)
						}
						return
					}
					
					waited := time.Since(waitStart)
					
					mu.Lock()
					totalRequests++
					waitTime += waited
					reqNum := totalRequests
					mu.Unlock()
					
					// Make a safe API call that won't modify anything
					// Here we're just getting the local endpoint ID
					start := time.Now()
					_, err = client.GetEndpointByDisplayName(ctx, "My Computer")
					callDuration := time.Since(start)
					
					mu.Lock()
					if err != nil {
						errorCount++
						fmt.Printf("[%d] Error: %v\n", reqNum, err)
					} else {
						successCount++
					}
					mu.Unlock()
					
					// Process rate limit headers from response
					// This would normally update the rate limiter from the response
					// ratelimit.UpdateRateLimiterFromResponse(limiter, resp)
					
					// Print periodic status
					if reqNum%10 == 0 || err != nil {
						elapsed := time.Since(startTime)
						actualRate := float64(reqNum) / elapsed.Seconds()
						avgWait := float64(waitTime) / float64(totalRequests) / float64(time.Millisecond)
						
						mu.Lock()
						fmt.Printf("[%s] Requests: %d, Success: %d, Errors: %d, Rate: %.1f req/sec, Avg Wait: %.1f ms, Last Call: %s\n",
							formatDuration(elapsed), reqNum, successCount, errorCount, 
							actualRate, avgWait, callDuration)
						
						// Print rate limiter stats
						stats := limiter.GetStats()
						fmt.Printf("  Rate Limiter: Limit=%.1f, Remaining=%.1f, Throttled=%d\n",
							stats.CurrentLimit, stats.RemainingTokens, stats.TotalThrottled)
						mu.Unlock()
					}
				}
			}
		}(i)
	}
	
	// Wait for completion
	wg.Wait()
	
	// Print final statistics
	elapsed := time.Since(startTime)
	actualRate := float64(totalRequests) / elapsed.Seconds()
	avgWait := float64(waitTime) / float64(totalRequests) / float64(time.Millisecond)
	
	fmt.Println("\n=== Final Statistics ===")
	fmt.Printf("Total Duration: %s\n", formatDuration(elapsed))
	fmt.Printf("Total Requests: %d\n", totalRequests)
	fmt.Printf("Successful Requests: %d (%.1f%%)\n", 
		successCount, float64(successCount)/float64(totalRequests)*100)
	fmt.Printf("Failed Requests: %d (%.1f%%)\n", 
		errorCount, float64(errorCount)/float64(totalRequests)*100)
	fmt.Printf("Average Rate: %.2f req/sec (Target: %.1f)\n", actualRate, reqPerSec)
	fmt.Printf("Average Wait Time: %.2f ms\n", avgWait)
	
	limiterStats := limiter.GetStats()
	fmt.Printf("Rate Limiter Throttled: %d\n", limiterStats.TotalThrottled)
	fmt.Printf("Total Wait Time: %s\n", limiterStats.TotalWaitTime)
}

// runBackoffTest demonstrates the exponential backoff with a simulated API
func runBackoffTest() {
	fmt.Println("=== Running Backoff Strategy Test ===")
	
	// Create backoff strategy
	backoff := ratelimit.NewExponentialBackoff(
		100*time.Millisecond,
		5*time.Second,
		2.0,
		5,
	)
	
	// Create a context
	ctx := context.Background()
	
	// Simulate an API call that fails a few times then succeeds
	attempt := 0
	maxFailures := 3
	
	err := ratelimit.RetryWithBackoff(ctx, func(ctx context.Context) error {
		attempt++
		fmt.Printf("Attempt %d/%d...\n", attempt, maxFailures+1)
		
		// Simulate work
		time.Sleep(50 * time.Millisecond)
		
		// Fail the first few attempts
		if attempt <= maxFailures {
			fmt.Printf("  Failed with temporary error\n")
			return errors.New("temporary error: service unavailable")
		}
		
		// Succeed on the last attempt
		fmt.Printf("  Success!\n")
		return nil
		
	}, backoff, ratelimit.IsRetryableError)
	
	if err != nil {
		fmt.Printf("Final result: Error - %v\n", err)
	} else {
		fmt.Printf("Final result: Success after %d attempts\n", attempt)
	}
	
	// Demonstrate non-retryable errors
	fmt.Println("\n=== Testing Non-Retryable Errors ===")
	
	err = ratelimit.RetryWithBackoff(ctx, func(ctx context.Context) error {
		fmt.Println("Attempt with non-retryable error...")
		return errors.New("permanent error: not found")
	}, backoff, ratelimit.IsRetryableError)
	
	if err != nil {
		fmt.Printf("Final result: Error - %v (not retried as expected)\n", err)
	}
	
	// Demonstrate context cancellation
	fmt.Println("\n=== Testing Context Cancellation ===")
	
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()
	
	attempt = 0
	start := time.Now()
	
	err = ratelimit.RetryWithBackoff(ctxWithTimeout, func(ctx context.Context) error {
		attempt++
		fmt.Printf("Attempt %d...\n", attempt)
		
		// Simulate work that takes time
		time.Sleep(100 * time.Millisecond)
		
		// Always fail to force retries
		return errors.New("temporary error: please retry")
		
	}, backoff, ratelimit.IsRetryableError)
	
	elapsed := time.Since(start)
	
	if errors.Is(err, context.DeadlineExceeded) {
		fmt.Printf("Final result: Context deadline exceeded after %s and %d attempts (as expected)\n", 
			elapsed, attempt)
	} else {
		fmt.Printf("Final result: Unexpected result - %v\n", err)
	}
}

// runCircuitBreakerTest demonstrates the circuit breaker pattern
func runCircuitBreakerTest() {
	fmt.Println("=== Running Circuit Breaker Test ===")
	
	// Create circuit breaker with visible state changes
	options := &ratelimit.CircuitBreakerOptions{
		Threshold:         3,                 // Open after 3 failures
		Timeout:           2 * time.Second,   // Half-open after 2 seconds
		HalfOpenSuccesses: 2,                 // Close after 2 successes
		OnStateChange: func(from, to ratelimit.CircuitBreakerState) {
			stateNames := map[ratelimit.CircuitBreakerState]string{
				ratelimit.CircuitClosed:   "CLOSED",
				ratelimit.CircuitOpen:     "OPEN",
				ratelimit.CircuitHalfOpen: "HALF-OPEN",
			}
			fmt.Printf("\n*** Circuit state changed: %s -> %s ***\n\n", 
				stateNames[from], stateNames[to])
		},
	}
	
	cb := ratelimit.NewCircuitBreaker(options)
	
	// Phase 1: Normal operation
	fmt.Println("Phase 1: Normal operation")
	ctx := context.Background()
	
	for i := 0; i < 5; i++ {
		err := cb.Execute(ctx, func(ctx context.Context) error {
			fmt.Printf("Request %d: Normal operation (success)\n", i+1)
			return nil
		})
		
		if err != nil {
			fmt.Printf("Unexpected error: %v\n", err)
		}
		
		time.Sleep(100 * time.Millisecond)
	}
	
	// Phase 2: Generate failures to trip circuit breaker
	fmt.Println("\nPhase 2: Generating failures to trip circuit breaker")
	
	// Make requests that fail
	for i := 0; i < 5; i++ {
		err := cb.Execute(ctx, func(ctx context.Context) error {
			fmt.Printf("Request %d: Simulating failure\n", i+1)
			return errors.New("simulated error")
		})
		
		if errors.Is(err, ratelimit.ErrCircuitOpen) {
			fmt.Printf("Circuit is open, request rejected\n")
		}
		
		time.Sleep(100 * time.Millisecond)
	}
	
	// Phase 3: Circuit is open, requests should fail fast
	fmt.Println("\nPhase 3: Circuit is open, requests should fail fast")
	
	for i := 0; i < 3; i++ {
		start := time.Now()
		
		err := cb.Execute(ctx, func(ctx context.Context) error {
			// This should not execute when circuit is open
			fmt.Println("This should not be printed when circuit is open")
			time.Sleep(500 * time.Millisecond)
			return nil
		})
		
		elapsed := time.Since(start)
		
		if errors.Is(err, ratelimit.ErrCircuitOpen) {
			fmt.Printf("Request %d: Correctly rejected (circuit open), took %s\n", 
				i+1, elapsed)
		} else if err != nil {
			fmt.Printf("Request %d: Unexpected error: %v\n", i+1, err)
		} else {
			fmt.Printf("Request %d: Unexpected success\n", i+1)
		}
		
		time.Sleep(100 * time.Millisecond)
	}
	
	// Phase 4: Wait for timeout and test half-open state
	fmt.Println("\nPhase 4: Waiting for timeout period (2 seconds)...")
	time.Sleep(2100 * time.Millisecond)
	
	// First request in half-open state succeeds
	fmt.Println("\nPhase 5: Testing half-open state")
	
	err := cb.Execute(ctx, func(ctx context.Context) error {
		fmt.Println("First request in half-open state (success)")
		return nil
	})
	
	if err != nil {
		fmt.Printf("Unexpected error: %v\n", err)
	}
	
	// Second request in half-open state succeeds (should close circuit)
	err = cb.Execute(ctx, func(ctx context.Context) error {
		fmt.Println("Second request in half-open state (success)")
		return nil
	})
	
	if err != nil {
		fmt.Printf("Unexpected error: %v\n", err)
	}
	
	// Phase 6: Circuit should be closed again
	fmt.Println("\nPhase 6: Circuit should be closed again, normal operation")
	
	for i := 0; i < 3; i++ {
		err := cb.Execute(ctx, func(ctx context.Context) error {
			fmt.Printf("Request %d: Normal operation after recovery\n", i+1)
			return nil
		})
		
		if err != nil {
			fmt.Printf("Unexpected error: %v\n", err)
		}
		
		time.Sleep(100 * time.Millisecond)
	}
}

// getAccessToken obtains a Globus access token using the auth flow
func getAccessToken() (string, error) {
	// Get client credentials from environment
	clientID := os.Getenv("GLOBUS_CLIENT_ID")
	clientSecret := os.Getenv("GLOBUS_CLIENT_SECRET")
	
	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("GLOBUS_CLIENT_ID and GLOBUS_CLIENT_SECRET environment variables are required")
	}
	
	// Create auth client
	authClient := auth.NewClient(clientID, clientSecret)
	authClient.SetRedirectURL("http://localhost:8080/callback")
	
	// Get authorization URL
	state := "ratelimit-demo-state"
	authURL := authClient.GetAuthorizationURL(
		state,
		auth.ScopeOpenID,
		auth.ScopeProfile,
		auth.ScopeEmail,
		transfer.TransferScope,
	)
	
	// Instruct user to visit the URL
	fmt.Printf("Please visit the following URL to authorize this application:\n\n%s\n\n", authURL)
	fmt.Printf("After authorization, you will be redirected to a localhost URL.\n")
	fmt.Printf("Copy the authorization code from the URL and paste it here: ")
	
	// Read authorization code
	var code string
	fmt.Scanln(&code)
	
	// Exchange code for tokens
	ctx := context.Background()
	tokenResponse, err := authClient.ExchangeAuthorizationCode(ctx, code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange authorization code: %w", err)
	}
	
	return tokenResponse.AccessToken, nil
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	seconds := int(d.Seconds())
	minutes := seconds / 60
	hours := minutes / 60
	
	seconds %= 60
	minutes %= 60
	
	if hours > 0 {
		return fmt.Sprintf("%dh%02dm%02ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm%02ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
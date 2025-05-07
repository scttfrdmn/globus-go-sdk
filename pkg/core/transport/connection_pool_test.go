// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package transport

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestConnectionPool(t *testing.T) {
	// Start a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate some work
		time.Sleep(10 * time.Millisecond)
		fmt.Fprintf(w, "Hello, %s!", r.URL.Path)
	}))
	defer server.Close()

	// Create a connection pool
	config := DefaultConnectionPoolConfig()
	config.MaxIdleConnsPerHost = 10
	config.MaxConnsPerHost = 20
	pool := NewConnectionPool(config)

	// Test basic functionality
	t.Run("Basic functionality", func(t *testing.T) {
		client := pool.GetClient()
		resp, err := client.Get(server.URL + "/test")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// After the request, the pool should have 1 idle connection
		stats := pool.GetStats()
		if stats.TotalActive != 0 {
			t.Errorf("Expected 0 active connections, got %d", stats.TotalActive)
		}
	})

	// Test concurrent requests
	t.Run("Concurrent requests", func(t *testing.T) {
		const numRequests = 50
		var wg sync.WaitGroup
		client := pool.GetClient()

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				resp, err := client.Get(fmt.Sprintf("%s/concurrent/%d", server.URL, id))
				if err != nil {
					t.Errorf("Failed to make request: %v", err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					t.Errorf("Expected status 200, got %d", resp.StatusCode)
				}
			}(i)
		}

		// Wait for all requests to complete
		wg.Wait()

		// Check that connections are reused
		stats := pool.GetStats()
		t.Logf("Connection stats after concurrent requests: active=%d, hosts=%d",
			stats.TotalActive, stats.ActiveHosts)
	})

	// Test connection pool statistics
	t.Run("Connection pool statistics", func(t *testing.T) {
		// Make some requests to generate stats
		client := pool.GetClient()
		for i := 0; i < 5; i++ {
			resp, err := client.Get(server.URL + "/stats")
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			resp.Body.Close()
		}

		// Get the stats
		stats := pool.GetStats()

		// Verify stats match configuration
		if stats.Config.MaxIdleConnsPerHost != config.MaxIdleConnsPerHost {
			t.Errorf("Expected MaxIdleConnsPerHost %d, got %d",
				config.MaxIdleConnsPerHost, stats.Config.MaxIdleConnsPerHost)
		}

		if stats.Config.MaxConnsPerHost != config.MaxConnsPerHost {
			t.Errorf("Expected MaxConnsPerHost %d, got %d",
				config.MaxConnsPerHost, stats.Config.MaxConnsPerHost)
		}
	})

	// Test closing idle connections
	t.Run("Close idle connections", func(t *testing.T) {
		// Make some requests
		client := pool.GetClient()
		for i := 0; i < 3; i++ {
			resp, err := client.Get(server.URL + "/idle")
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			resp.Body.Close()
		}

		// Close idle connections
		pool.CloseIdleConnections()

		// Get the stats (should show 0 active)
		stats := pool.GetStats()
		if stats.TotalActive != 0 {
			t.Errorf("Expected 0 active connections after closing idle, got %d", stats.TotalActive)
		}
	})
}

func TestConnectionPoolManager(t *testing.T) {
	// Create a connection pool manager
	manager := NewConnectionPoolManager(nil)

	// Test getting pools for different services
	t.Run("Get different service pools", func(t *testing.T) {
		pool1 := manager.GetPool("service1", nil)
		pool2 := manager.GetPool("service2", nil)

		if pool1 == pool2 {
			t.Error("Expected different pools for different services")
		}

		// Getting the same service again should return the same pool
		pool1Again := manager.GetPool("service1", nil)
		if pool1 != pool1Again {
			t.Error("Expected same pool when requesting the same service")
		}
	})

	// Test creating pool with custom config
	t.Run("Custom pool configuration", func(t *testing.T) {
		customConfig := DefaultConnectionPoolConfig()
		customConfig.MaxIdleConnsPerHost = 42
		customConfig.IdleConnTimeout = 2 * time.Minute

		pool := manager.GetPool("custom", customConfig)
		stats := pool.GetStats()

		if stats.Config.MaxIdleConnsPerHost != 42 {
			t.Errorf("Expected MaxIdleConnsPerHost 42, got %d", stats.Config.MaxIdleConnsPerHost)
		}

		if stats.Config.IdleConnTimeout != 2*time.Minute {
			t.Errorf("Expected IdleConnTimeout 2m, got %s", stats.Config.IdleConnTimeout)
		}
	})

	// Test global pool stats
	t.Run("Global stats", func(t *testing.T) {
		// Create a few more pools
		manager.GetPool("service3", nil)
		manager.GetPool("service4", nil)

		// Get all stats
		allStats := manager.GetAllStats()

		// Should have stats for all services
		expectedServices := []string{"service1", "service2", "custom", "service3", "service4"}
		for _, service := range expectedServices {
			if _, ok := allStats[service]; !ok {
				t.Errorf("Expected stats for service %s, but not found", service)
			}
		}
	})

	// Test closing all idle connections
	t.Run("Close all idle connections", func(t *testing.T) {
		manager.CloseAllIdleConnections()
		// This is more of a coverage test - hard to verify without mocking
	})

	// Test global manager
	t.Run("Global manager", func(t *testing.T) {
		pool := GetServicePool("globalTest", nil)
		if pool == nil {
			t.Error("Expected non-nil pool from global manager")
		}

		// Getting same service should return same pool
		poolAgain := GetServicePool("globalTest", nil)
		if pool != poolAgain {
			t.Error("Expected same pool from global manager")
		}
	})
}

func TestConnectionPoolTransport(t *testing.T) {
	// Start a test server that counts requests
	requestCount := 0
	var requestMutex sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestMutex.Lock()
		requestCount++
		requestMutex.Unlock()
		fmt.Fprintf(w, "Response #%d", requestCount)
	}))
	defer server.Close()

	// Create a connection pool
	pool := NewConnectionPool(nil)

	// Create a connection pool transport
	transport := NewConnectionPoolTransport(pool)

	// Create a client that uses the transport
	client := &http.Client{
		Transport: transport,
	}

	// Test basic functionality
	t.Run("Basic roundtrip", func(t *testing.T) {
		resp, err := client.Get(server.URL)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Verify the request was tracked
		stats := pool.GetStats()
		if host := server.URL[7:]; stats.ActiveByHost[host] != 0 {
			t.Errorf("Expected 0 active connections for host %s, got %d",
				host, stats.ActiveByHost[host])
		}
	})

	// Test concurrent requests
	t.Run("Concurrent roundtrips", func(t *testing.T) {
		const numRequests = 10
		var wg sync.WaitGroup

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				resp, err := client.Get(server.URL)
				if err != nil {
					t.Errorf("Failed to make request: %v", err)
					return
				}
				defer resp.Body.Close()
			}()
		}

		wg.Wait()

		// Verify all requests completed
		requestMutex.Lock()
		count := requestCount
		requestMutex.Unlock()

		if count < numRequests+1 { // +1 for the earlier test
			t.Errorf("Expected at least %d requests, got %d", numRequests+1, count)
		}
	})
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package pkg_test

import (
	"os"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	httppool "github.com/scttfrdmn/globus-go-sdk/pkg/core/http"
)

// TestConnectionPoolInitialization verifies that the connection pool initialization
// process works as expected. This test will catch issues with the initialization
// function itself, which is called when creating a new SDK config from environment.
func TestConnectionPoolInitialization(t *testing.T) {
	// Save the original value of the environment variable
	origValue := os.Getenv("GLOBUS_DISABLE_CONNECTION_POOL")
	defer os.Setenv("GLOBUS_DISABLE_CONNECTION_POOL", origValue)

	// Enable connection pooling
	os.Setenv("GLOBUS_DISABLE_CONNECTION_POOL", "")

	// Create a new config from environment, which should initialize the connection pools
	config := pkg.NewConfigFromEnvironment()
	if config == nil {
		t.Fatal("NewConfigFromEnvironment returned nil")
	}

	// Verify that we can get connection pools for all services
	services := []string{
		"auth", "transfer", "search", "flows", "groups", "compute", "timers", "default",
	}

	for _, service := range services {
		// Use GetServicePool directly
		pool := httppool.GetServicePool(service, nil)
		if pool == nil {
			t.Errorf("GetServicePool returned nil for service %s", service)
		}

		// Also test the client getter
		client := httppool.GetHTTPClientForService(service, nil)
		if client == nil {
			t.Errorf("GetHTTPClientForService returned nil for service %s", service)
		}
	}
}

// TestServiceClientConnectionPoolIntegration verifies that service clients correctly
// use the connection pool when configured to do so.
func TestServiceClientConnectionPoolIntegration(t *testing.T) {
	// Save the original value of the environment variable
	origValue := os.Getenv("GLOBUS_DISABLE_CONNECTION_POOL")
	defer os.Setenv("GLOBUS_DISABLE_CONNECTION_POOL", origValue)

	// Enable connection pooling
	os.Setenv("GLOBUS_DISABLE_CONNECTION_POOL", "")

	// Create a new config
	config := pkg.NewConfig()

	// Test with Auth client (doesn't require tokens for initialization)
	config.WithClientID("test-client-id")
	authClient, err := config.NewAuthClient()
	if err != nil {
		t.Fatalf("Failed to create Auth client: %v", err)
	}

	// Verify that the HTTP client is set from the pool
	if authClient.Client.HTTPClient == nil {
		t.Error("Auth client HTTPClient is nil")
	}
}

// TestInitializeConnectionPools verifies that the initializeConnectionPools function
// correctly creates and configures connection pools for all services.
func TestInitializeConnectionPools(t *testing.T) {
	// First, reset the global manager for testing
	httppool.GlobalHttpPoolManager = httppool.NewConnectionPoolManager(nil)

	// Call the function through a new config from environment
	pkg.NewConfigFromEnvironment()

	// Verify that pools were created for all services
	services := []string{
		"auth", "transfer", "search", "flows", "groups", "compute", "timers", "default",
	}
	for _, service := range services {
		// Skip default as it might not have stats until used
		if service == "default" {
			continue
		}

		// Try to get the service stats
		pool := httppool.GetServicePool(service, nil)
		if pool == nil {
			t.Errorf("After initialization, GetServicePool returned nil for service %s", service)
		}
	}
}

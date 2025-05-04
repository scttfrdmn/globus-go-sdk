// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
	"os"
	"time"
	
	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/http"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

// This example demonstrates the connection pooling functionality
// of the Globus Go SDK. It shows how to:
// 1. Use the default connection pooling
// 2. Configure custom connection pools
// 3. Monitor connection pool statistics
func main() {
	// Check if access token is provided
	accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
	if accessToken == "" {
		fmt.Println("Please set GLOBUS_ACCESS_TOKEN environment variable")
		os.Exit(1)
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	
	fmt.Println("Connection Pooling Example")
	fmt.Println("==========================")
	
	// 1. Using the default connection pooling
	usingDefaultPool()
	
	// 2. Creating custom connection pools
	usingCustomPool(ctx, accessToken)
	
	// 3. Monitoring pool statistics
	monitorPoolStats(ctx, accessToken)
}

// usingDefaultPool demonstrates how to use the default connection pooling
func usingDefaultPool() {
	fmt.Println("\n1. Using Default Connection Pool")
	fmt.Println("-------------------------------")
	
	// Connection pooling is enabled by default when creating a config from environment
	config := pkg.NewConfigFromEnvironment()
	
	// Create multiple service clients
	authClient, err := config.NewAuthClient()
	if err != nil {
		fmt.Printf("Failed to create auth client: %v\n", err)
		return
	}
	transferClient, err := config.NewTransferClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
	if err != nil {
		fmt.Printf("Failed to create transfer client: %v\n", err)
		return
	}
	searchClient, err := config.NewSearchClient(os.Getenv("GLOBUS_ACCESS_TOKEN"))
	if err != nil {
		fmt.Printf("Failed to create search client: %v\n", err)
		return
	}
	
	// The clients now share connection pools based on service type
	fmt.Println("Created Auth, Transfer, and Search clients with connection pooling")
	fmt.Println("Each service uses its own optimized connection pool")
	fmt.Println("Pool settings are based on expected usage patterns for each service type")
}

// usingCustomPool demonstrates how to create and use custom connection pools
func usingCustomPool(ctx context.Context, accessToken string) {
	fmt.Println("\n2. Using Custom Connection Pools")
	fmt.Println("------------------------------")
	
	// Create a custom connection pool configuration
	customConfig := &http.ConnectionPoolConfig{
		MaxIdleConnsPerHost:   12,  // More idle connections per host
		MaxConnsPerHost:       24,  // Higher connection limit
		IdleConnTimeout:       30 * time.Second, // Shorter idle timeout
		ResponseHeaderTimeout: 10 * time.Second, // Faster header timeout
	}
	
	// Register the custom pool with a service name
	customPool := http.GetServicePool("custom-transfer", customConfig)
	
	// Create an HTTP client using the custom pool
	httpClient := customPool.GetClient()
	
	// Create an SDK config with the custom client
	sdkConfig := pkg.NewConfig().
		WithClientID(os.Getenv("GLOBUS_CLIENT_ID")).
		WithClientSecret(os.Getenv("GLOBUS_CLIENT_SECRET"))
	
	// Create a Transfer client
	transferClient, err := sdkConfig.NewTransferClient(accessToken)
	if err != nil {
		fmt.Printf("Failed to create transfer client: %v\n", err)
		return
	}
	
	// Override the HTTP client with our custom pooled client
	transferClient.Client.HTTPClient = httpClient
	
	// Use the client with custom connection pool
	fmt.Println("Created Transfer client with custom connection pool settings:")
	fmt.Printf("  MaxIdleConnsPerHost: %d\n", customConfig.MaxIdleConnsPerHost)
	fmt.Printf("  MaxConnsPerHost:     %d\n", customConfig.MaxConnsPerHost)
	fmt.Printf("  IdleConnTimeout:     %s\n", customConfig.IdleConnTimeout)
	
	// Try listing endpoints
	listEndpoints(ctx, transferClient)
}

// monitorPoolStats demonstrates how to monitor connection pool statistics
func monitorPoolStats(ctx context.Context, accessToken string) {
	fmt.Println("\n3. Monitoring Connection Pool Statistics")
	fmt.Println("--------------------------------------")
	
	// Create SDK config
	config := pkg.NewConfigFromEnvironment()
	
	// Create a Transfer client
	transferClient, err := config.NewTransferClient(accessToken)
	if err != nil {
		fmt.Printf("Failed to create transfer client: %v\n", err)
		return
	}
	
	// Make some requests to generate connections
	for i := 0; i < 5; i++ {
		listEndpoints(ctx, transferClient)
		// Small delay to allow connections to be established
		time.Sleep(200 * time.Millisecond)
	}
	
	// Get pool statistics
	stats := http.GetServicePool("transfer", nil).GetStats()
	
	// Display pool statistics
	fmt.Println("Connection Pool Statistics for Transfer Service:")
	fmt.Printf("  Active Hosts:           %d\n", stats.ActiveHosts)
	fmt.Printf("  Total Active:           %d\n", stats.TotalActive)
	fmt.Printf("  Max Idle Conns/Host:    %d\n", stats.Config.MaxIdleConnsPerHost)
	fmt.Printf("  Max Conns/Host:         %d\n", stats.Config.MaxConnsPerHost)
	fmt.Printf("  Idle Conn Timeout:      %s\n", stats.Config.IdleConnTimeout)
	
	// Get global statistics
	allStats := http.GlobalHttpPoolManager.GetAllStats()
	fmt.Println("\nAll Connection Pools:")
	for service, stat := range allStats {
		fmt.Printf("  %s: %d active connection(s)\n", service, stat.TotalActive)
	}
}

// Helper function to list endpoints
func listEndpoints(ctx context.Context, client *transfer.Client) {
	endpoints, err := client.ListEndpoints(ctx, nil)
	if err != nil {
		fmt.Printf("Error listing endpoints: %v\n", err)
		return
	}
	fmt.Printf("Found %d endpoints\n", len(endpoints.Data))
}
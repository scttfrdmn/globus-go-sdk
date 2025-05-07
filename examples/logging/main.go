// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/logging"
)

func main() {
	// Check for access token
	accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
	if accessToken == "" {
		fmt.Println("Please set GLOBUS_ACCESS_TOKEN environment variable")
		os.Exit(1)
	}

	// Demonstrate text logging
	fmt.Println("\n=== Text Logging Example ===")
	textExample(accessToken)

	// Demonstrate JSON logging
	fmt.Println("\n=== JSON Logging Example ===")
	jsonExample(accessToken)

	// Demonstrate tracing
	fmt.Println("\n=== Tracing Example ===")
	tracingExample(accessToken)
}

func textExample(accessToken string) {
	// Create a configuration with text logging
	config := pkg.NewConfigFromEnvironment().
		WithClientOption(logging.WithLogLevel(logging.LogLevelDebug))

	// Create a client
	transferClient := config.NewTransferClient(accessToken)

	// Context for the operation
	ctx := context.Background()

	// List endpoints (this will generate logs in text format)
	endpoints, err := transferClient.ListEndpoints(ctx, nil)
	if err != nil {
		fmt.Printf("Error listing endpoints: %v\n", err)
		return
	}

	// Display the results
	fmt.Printf("Found %d endpoints\n", len(endpoints.Data))
}

func jsonExample(accessToken string) {
	// Create a configuration with JSON logging
	config := pkg.NewConfigFromEnvironment().
		WithClientOption(logging.WithLogLevel(logging.LogLevelDebug)).
		WithClientOption(logging.WithJSONLogging())

	// Create a client
	transferClient := config.NewTransferClient(accessToken)

	// Context for the operation
	ctx := context.Background()

	// List endpoints (this will generate logs in JSON format)
	endpoints, err := transferClient.ListEndpoints(ctx, nil)
	if err != nil {
		fmt.Printf("Error listing endpoints: %v\n", err)
		return
	}

	// Display the results
	fmt.Printf("Found %d endpoints\n", len(endpoints.Data))
}

func tracingExample(accessToken string) {
	// Create a configuration with tracing enabled
	config := pkg.NewConfigFromEnvironment().
		WithClientOption(logging.WithLogLevel(logging.LogLevelTrace)).
		WithClientOption(logging.WithTracing("example-trace-id"))

	// Create a client
	transferClient := config.NewTransferClient(accessToken)

	// Context for the operation
	ctx := context.Background()

	// List endpoints (this will generate detailed request/response logs)
	endpoints, err := transferClient.ListEndpoints(ctx, nil)
	if err != nil {
		fmt.Printf("Error listing endpoints: %v\n", err)
		return
	}

	// Display the results
	fmt.Printf("Found %d endpoints\n", len(endpoints.Data))

	// You can now trace the entire operation using the trace ID "example-trace-id"
	// In a real application, you might pass this trace ID to other components
	fmt.Println("Trace ID: example-trace-id")
}

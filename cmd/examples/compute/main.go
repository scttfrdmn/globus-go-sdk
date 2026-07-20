// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025-2026 Scott Friedman and Project Contributors
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/scttfrdmn/globus-go-sdk/v3/pkg"
	"github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/compute"
)

// The Globus Compute API takes and returns open-ended documents; the Go client
// mirrors that with map[string]interface{} bodies and results.
func main() {
	accessToken := os.Getenv("GLOBUS_ACCESS_TOKEN")
	if accessToken == "" {
		fmt.Println("Please set the GLOBUS_ACCESS_TOKEN environment variable")
		os.Exit(1)
	}

	config := pkg.NewConfigFromEnvironment()
	computeClient, err := config.NewComputeClient(accessToken)
	if err != nil {
		log.Fatalf("Failed to create compute client: %v", err)
	}
	ctx := context.Background()

	// Service version. With no service argument the API returns a bare string;
	// GetVersion returns it as an untyped value.
	if version, err := computeClient.GetVersion(ctx, ""); err == nil {
		fmt.Printf("Compute service version: %v\n", version)
	}

	// List endpoints the caller owns.
	endpoints, err := computeClient.GetEndpoints(ctx, &compute.GetEndpointsOptions{Role: "owner"})
	if err != nil {
		log.Printf("Failed to list endpoints: %v", err)
	} else {
		b, _ := json.MarshalIndent(endpoints, "", "  ")
		fmt.Printf("Endpoints: %s\n", b)
	}

	// Register a function (passthrough document).
	fn, err := computeClient.RegisterFunction(ctx, map[string]interface{}{
		"function_name": "hello",
		"function_code": "def hello(name='World'):\n    return f'Hello, {name}!'\n",
	})
	if err != nil {
		log.Printf("Failed to register function: %v", err)
		return
	}
	functionID, _ := fn["function_id"].(string)
	fmt.Printf("Registered function: %s\n", functionID)

	// Clean up.
	if functionID != "" {
		if _, err := computeClient.DeleteFunction(ctx, functionID); err != nil {
			log.Printf("Failed to delete function: %v", err)
		}
	}
	fmt.Println("Done.")
}

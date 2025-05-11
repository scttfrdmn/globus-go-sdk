// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package main

import (
	"fmt"
	"log"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/pool"
)

// This script simulates how a downstream project would use the SDK.
// It specifically tests the code path that was failing due to missing functions.

func main() {
	fmt.Println("Simulating downstream project that imports the SDK...")

	// Initialize the SDK in the same way a downstream project would
	fmt.Println("Step 1: Setting up connection pool manager...")
	manager := pool.NewPoolManager(nil)
	if manager == nil {
		log.Fatal("Failed to create pool manager")
	}

	// This is the function that was missing in the original issue
	fmt.Println("Step 2: Setting connection pool manager...")
	core.SetConnectionPoolManager(manager)

	// This is the second function that was missing in the original issue
	fmt.Println("Step 3: Enabling default connection pool...")
	core.EnableDefaultConnectionPool()

	// Test getting a connection pool
	fmt.Println("Step 4: Getting connection pool for a service...")
	p := core.GetConnectionPool("auth", nil)
	if p == nil {
		log.Fatal("Failed to get connection pool")
	}

	// Test getting an HTTP client
	fmt.Println("Step 5: Getting HTTP client for a service...")
	client := core.GetHTTPClientForService("auth")
	if client == nil {
		log.Fatal("Failed to get HTTP client")
	}

	fmt.Println("\n✅ SUCCESS: All steps completed successfully!")
	fmt.Println("The fix for issue #13 works correctly in a downstream project scenario.")
}

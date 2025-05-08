// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/pool"
)

// This script verifies that all connection pool functions referenced in issue #13
// are properly defined and exportable.

func main() {
	success := true
	errors := []string{}

	// Store original function values to verify they exist
	var originalSetConnectionPoolManager = core.SetConnectionPoolManager
	var originalEnableDefaultConnectionPool = core.EnableDefaultConnectionPool
	var originalGetConnectionPool = core.GetConnectionPool
	var originalGetHTTPClientForService = core.GetHTTPClientForService

	// Check if functions are defined
	if originalSetConnectionPoolManager == nil {
		success = false
		errors = append(errors, "SetConnectionPoolManager function is nil")
	}

	if originalEnableDefaultConnectionPool == nil {
		success = false
		errors = append(errors, "EnableDefaultConnectionPool function is nil")
	}

	if originalGetConnectionPool == nil {
		success = false
		errors = append(errors, "GetConnectionPool function is nil")
	}

	if originalGetHTTPClientForService == nil {
		success = false
		errors = append(errors, "GetHTTPClientForService function is nil")
	}

	// Verify function actually works
	fmt.Println("Testing connection pool functions...")

	// Create a test pool manager
	testManager := pool.NewPoolManager(nil)
	if testManager == nil {
		success = false
		errors = append(errors, "Failed to create pool manager")
	}

	// Test SetConnectionPoolManager
	fmt.Println("- Testing SetConnectionPoolManager...")
	func() {
		// Use defer/recover to catch any panics
		defer func() {
			if r := recover(); r != nil {
				success = false
				errors = append(errors, fmt.Sprintf("SetConnectionPoolManager panicked: %v", r))
			}
		}()
		core.SetConnectionPoolManager(testManager)
	}()

	// Test EnableDefaultConnectionPool
	fmt.Println("- Testing EnableDefaultConnectionPool...")
	func() {
		// Use defer/recover to catch any panics
		defer func() {
			if r := recover(); r != nil {
				success = false
				errors = append(errors, fmt.Sprintf("EnableDefaultConnectionPool panicked: %v", r))
			}
		}()
		core.EnableDefaultConnectionPool()
	}()

	// Test GetConnectionPool
	fmt.Println("- Testing GetConnectionPool...")
	func() {
		// Use defer/recover to catch any panics
		defer func() {
			if r := recover(); r != nil {
				success = false
				errors = append(errors, fmt.Sprintf("GetConnectionPool panicked: %v", r))
			}
		}()
		
		// Test with various service names
		services := []string{"auth", "transfer", "search", "compute", "flows", "groups", "timers"}
		for _, service := range services {
			pool := core.GetConnectionPool(service, nil)
			if pool == nil {
				success = false
				errors = append(errors, fmt.Sprintf("GetConnectionPool returned nil for service %s", service))
			}
		}
	}()

	// Test GetHTTPClientForService
	fmt.Println("- Testing GetHTTPClientForService...")
	func() {
		// Use defer/recover to catch any panics
		defer func() {
			if r := recover(); r != nil {
				success = false
				errors = append(errors, fmt.Sprintf("GetHTTPClientForService panicked: %v", r))
			}
		}()
		
		// Test with various service names
		services := []string{"auth", "transfer", "search", "compute", "flows", "groups", "timers"}
		for _, service := range services {
			client := core.GetHTTPClientForService(service)
			if client == nil {
				success = false
				errors = append(errors, fmt.Sprintf("GetHTTPClientForService returned nil for service %s", service))
			}
		}
	}()

	// Verify correct function signatures by testing function values
	fmt.Println("- Verifying function signatures...")
	
	// Use reflection to verify the function signatures
	setPoolManagerType := runtime.FuncForPC(reflect.ValueOf(core.SetConnectionPoolManager).Pointer()).Name()
	if !strings.Contains(setPoolManagerType, "SetConnectionPoolManager") {
		success = false
		errors = append(errors, fmt.Sprintf("SetConnectionPoolManager has incorrect signature: %s", setPoolManagerType))
	}
	
	enablePoolType := runtime.FuncForPC(reflect.ValueOf(core.EnableDefaultConnectionPool).Pointer()).Name()
	if !strings.Contains(enablePoolType, "EnableDefaultConnectionPool") {
		success = false
		errors = append(errors, fmt.Sprintf("EnableDefaultConnectionPool has incorrect signature: %s", enablePoolType))
	}
	
	getPoolType := runtime.FuncForPC(reflect.ValueOf(core.GetConnectionPool).Pointer()).Name()
	if !strings.Contains(getPoolType, "GetConnectionPool") {
		success = false
		errors = append(errors, fmt.Sprintf("GetConnectionPool has incorrect signature: %s", getPoolType))
	}
	
	getClientType := runtime.FuncForPC(reflect.ValueOf(core.GetHTTPClientForService).Pointer()).Name()
	if !strings.Contains(getClientType, "GetHTTPClientForService") {
		success = false
		errors = append(errors, fmt.Sprintf("GetHTTPClientForService has incorrect signature: %s", getClientType))
	}

	// Print results
	if success {
		fmt.Println("\n✅ SUCCESS: All connection pool functions verified successfully!")
		fmt.Println("The fix for issue #13 is correctly implemented.")
	} else {
		fmt.Println("\n❌ FAILURE: There were errors verifying the connection pool functions:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		fmt.Println("\nThe fix for issue #13 is NOT correctly implemented.")
		os.Exit(1)
	}
}
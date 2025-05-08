// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
//
// This standalone tool verifies that required exported symbols are available
// in the SDK packages. It's designed to catch issues like missing function
// exports that would only be noticed during compilation of dependent projects.
//
// Run with: go run scripts/verify_package_exports.go
package scripts

import (
	"fmt"
	"os"
	"reflect"

	// Import SDK packages that we want to verify
	"github.com/scttfrdmn/globus-go-sdk/pkg"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/http"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
)

// requiredExport defines a function or type that must be exported from a package
type requiredExport struct {
	name        string
	packagePath string
	value       interface{}
	isType      bool
}

// verifyExport checks if the given export is properly defined and not nil
func verifyExport(export requiredExport) error {
	if export.value == nil {
		return fmt.Errorf("%s.%s is nil", export.packagePath, export.name)
	}

	if export.isType {
		// For types, we verify that it exists but don't do nil checks
		return nil
	}

	// For functions, check if they're nil using reflection
	val := reflect.ValueOf(export.value)
	if val.IsNil() {
		return fmt.Errorf("%s.%s is defined but nil", export.packagePath, export.name)
	}

	return nil
}

// validateInterfaceImplementation is a commented-out function that we replaced
// with direct type assertions for better reliability
/*
func validateInterfaceImplementation(concreteName string, concrete, iface interface{}) error {
	// Use direct type assertion which is more reliable than reflection
	_, ok := concrete.(iface)
	if !ok {
		return fmt.Errorf("%s doesn't implement %s", concreteName,
			reflect.TypeOf(iface).Elem().Name())
	}
	return nil
}
*/

// Required for standalone execution
func main() {
	if VerifyPackageExports() {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

// VerifyPackageExports checks that all required exports are available
func VerifyPackageExports() bool {
	// Define the critical exports we need to verify
	criticalExports := []requiredExport{
		// HTTP Package exports - these were problematic in issue #11
		{"GetServicePool", "github.com/scttfrdmn/globus-go-sdk/pkg/core/http", http.GetServicePool, false},
		{"GetHTTPClientForService", "github.com/scttfrdmn/globus-go-sdk/pkg/core/http", http.GetHTTPClientForService, false},
		{"NewConnectionPool", "github.com/scttfrdmn/globus-go-sdk/pkg/core/http", http.NewConnectionPool, false},
		{"NewConnectionPoolManager", "github.com/scttfrdmn/globus-go-sdk/pkg/core/http", http.NewConnectionPoolManager, false},
		{"NewHttpConnectionPool", "github.com/scttfrdmn/globus-go-sdk/pkg/core/http", http.NewHttpConnectionPool, false},
		{"NewHttpConnectionPoolManager", "github.com/scttfrdmn/globus-go-sdk/pkg/core/http", http.NewHttpConnectionPoolManager, false},
		{"GlobalHttpPoolManager", "github.com/scttfrdmn/globus-go-sdk/pkg/core/http", http.GlobalHttpPoolManager, false},

		// SDK Package exports
		{"Version", "github.com/scttfrdmn/globus-go-sdk/pkg", pkg.Version, true},
		{"NewConfig", "github.com/scttfrdmn/globus-go-sdk/pkg", pkg.NewConfig, false},
		{"NewConfigFromEnvironment", "github.com/scttfrdmn/globus-go-sdk/pkg", pkg.NewConfigFromEnvironment, false},
		{"GetScopesByService", "github.com/scttfrdmn/globus-go-sdk/pkg", pkg.GetScopesByService, false},
	}

	// Check each export
	hasErrors := false
	for _, export := range criticalExports {
		if err := verifyExport(export); err != nil {
			fmt.Printf("ERROR: %v\n", err)
			hasErrors = true
		} else {
			fmt.Printf("✓ %s.%s is properly exported\n", export.packagePath, export.name)
		}
	}

	// Check interface implementations using direct type assertions
	// This is more reliable than using reflect.Implements() which can sometimes give false negatives
	poolInstance := http.NewHttpConnectionPool(nil)
	_, ok1 := interface{}(poolInstance).(interfaces.ConnectionPool)
	if !ok1 {
		fmt.Printf("ERROR: HttpConnectionPool doesn't implement interfaces.ConnectionPool\n")
		hasErrors = true
	} else {
		fmt.Printf("✓ HttpConnectionPool correctly implements interface\n")
	}

	managerInstance := http.NewHttpConnectionPoolManager(nil)
	_, ok2 := interface{}(managerInstance).(interfaces.ConnectionPoolManager)
	if !ok2 {
		fmt.Printf("ERROR: HttpConnectionPoolManager doesn't implement interfaces.ConnectionPoolManager\n")
		hasErrors = true
	} else {
		fmt.Printf("✓ HttpConnectionPoolManager correctly implements interface\n")
	}

	// Finally, verify that we can use the HTTP pool correctly
	pool := http.GetServicePool("test", nil)
	if pool == nil {
		fmt.Println("ERROR: GetServicePool returned nil")
		hasErrors = true
	} else {
		fmt.Println("✓ GetServicePool works correctly")
	}

	client := http.GetHTTPClientForService("test", nil)
	if client == nil {
		fmt.Println("ERROR: GetHTTPClientForService returned nil")
		hasErrors = true
	} else {
		fmt.Println("✓ GetHTTPClientForService works correctly")
	}

	// Summary
	if hasErrors {
		fmt.Println("\n❌ FAILED: Some exports are missing or invalid")
		fmt.Println("This will likely cause compilation errors in dependent projects")
		return false
	} else {
		fmt.Println("\n✅ SUCCESS: All required exports are available")
		return true
	}
}

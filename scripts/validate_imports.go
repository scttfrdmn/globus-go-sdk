// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package main

import (
	"fmt"
	"reflect"
	"runtime"

	// Import each package separately to check for import cycles
	"github.com/scttfrdmn/globus-go-sdk/pkg/core"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/pool"
	coreconfig "github.com/scttfrdmn/globus-go-sdk/pkg/core/config"
	coretransport "github.com/scttfrdmn/globus-go-sdk/pkg/core/transport"
)

// This script validates that there are no import cycles with our fix.
// It does this by importing each relevant package and performing a simple operation
// with each one to ensure they're properly loaded.

func main() {
	fmt.Println("Validating imports for all relevant packages...")
	
	// Validate core package
	fmt.Println("\nValidating pkg/core:")
	setFunc := reflect.ValueOf(core.SetConnectionPoolManager)
	setFuncName := runtime.FuncForPC(setFunc.Pointer()).Name()
	fmt.Println("✅ core.SetConnectionPoolManager exists:", setFuncName)
	
	// Validate interfaces package
	fmt.Println("\nValidating pkg/core/interfaces:")
	var cpConfig interfaces.ConnectionPoolConfig
	fmt.Println("✅ interfaces.ConnectionPoolConfig type exists:", reflect.TypeOf(cpConfig))
	
	// Validate pool package
	fmt.Println("\nValidating pkg/core/pool:")
	manager := pool.NewPoolManager(nil)
	fmt.Println("✅ pool.NewPoolManager works:", reflect.TypeOf(manager))
	
	// Validate config package
	fmt.Println("\nValidating pkg/core/config:")
	config := coreconfig.Config{}
	fmt.Println("✅ config.Config type exists:", reflect.TypeOf(config))
	
	// Validate transport package
	fmt.Println("\nValidating pkg/core/transport:")
	dtOptions := &coretransport.Options{Debug: true}
	fmt.Println("✅ transport.Options type exists:", reflect.TypeOf(dtOptions))
	
	fmt.Println("\n✅ SUCCESS: All packages imported successfully without import cycles!")
}
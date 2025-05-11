// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package deprecation_test

import (
	"fmt"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/deprecation"
)

// MockLogger implements the interfaces.Logger interface for example purposes
type MockLogger struct{}

func (l *MockLogger) Debug(format string, args ...interface{}) {
	fmt.Printf("[DEBUG] "+format+"\n", args...)
}

func (l *MockLogger) Info(format string, args ...interface{}) {
	fmt.Printf("[INFO] "+format+"\n", args...)
}

func (l *MockLogger) Warn(format string, args ...interface{}) {
	fmt.Printf("[WARN] "+format+"\n", args...)
}

func (l *MockLogger) Error(format string, args ...interface{}) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}

// Example demonstrates how to use the deprecation system in a function
func Example() {
	logger := &MockLogger{}

	// Example of a deprecated function that logs a warning
	deprecatedFunction := func() {
		// Log a deprecation warning
		deprecation.LogWarning(
			logger,
			"deprecatedFunction",
			"v0.9.0",
			"v1.0.0",
			"Use newFunction() instead.",
		)

		// Function implementation...
		fmt.Println("Doing deprecated function work")
	}

	// Call the deprecated function
	deprecatedFunction()

	// Create a feature info object for more complex cases
	uploadFeatureInfo := deprecation.CreateFeatureInfo(
		"UploadWithoutChecksum",
		"v0.9.0",
		"v1.0.0",
		"Use UploadWithChecksum() for better data integrity.",
	)

	// Example of using a FeatureInfo object
	uploadWithoutChecksum := func() {
		// Log the deprecation warning using the feature info
		deprecation.LogFeatureWarning(logger, uploadFeatureInfo)

		// Function implementation...
		fmt.Println("Uploading without checksum")
	}

	// Call the function that uses FeatureInfo
	uploadWithoutChecksum()

	// Since WarnOnce is true by default, a second call won't log a warning
	uploadWithoutChecksum()

	// Output:
	// [WARN] DEPRECATED: deprecatedFunction was deprecated in v0.9.0 and will be removed in v1.0.0. Use newFunction() instead.
	// Doing deprecated function work
	// [WARN] DEPRECATED: UploadWithoutChecksum was deprecated in v0.9.0 and will be removed in v1.0.0. Use UploadWithChecksum() for better data integrity.
	// Uploading without checksum
	// Uploading without checksum
}

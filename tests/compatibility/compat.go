// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors

// Package compatibility provides a framework for testing backward compatibility
// of the Globus Go SDK with dependent code.
package compatibility

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test represents a compatibility test
type Test interface {
	// Name returns the name of the test
	Name() string

	// Setup performs any necessary setup before the test
	Setup(ctx context.Context) error

	// Run runs the compatibility test
	Run(ctx context.Context, version string, t *testing.T) error

	// Teardown performs any necessary cleanup after the test
	Teardown(ctx context.Context) error
}

// TestRegistry maintains a registry of compatibility tests
type TestRegistry struct {
	tests []Test
}

// GlobalRegistry is the global registry of compatibility tests
var GlobalRegistry = &TestRegistry{}

// RegisterTest registers a compatibility test
func RegisterTest(test Test) {
	GlobalRegistry.tests = append(GlobalRegistry.tests, test)
}

// RunAllTests runs all registered compatibility tests
func (tr *TestRegistry) RunAllTests(ctx context.Context, t *testing.T) {
	// Get the target version from environment or use current
	version := os.Getenv("VERSION")
	if version == "" {
		version = "current"
	}

	t.Logf("Running compatibility tests for version: %s", version)

	// Run each test
	for _, test := range tr.tests {
		testName := test.Name()
		t.Run(testName, func(t *testing.T) {
			// Setup
			if err := test.Setup(ctx); err != nil {
				t.Fatalf("Failed to setup test %s: %v", testName, err)
			}

			// Run test
			if err := test.Run(ctx, version, t); err != nil {
				t.Errorf("Compatibility test %s failed: %v", testName, err)
			}

			// Teardown
			if err := test.Teardown(ctx); err != nil {
				t.Errorf("Failed to teardown test %s: %v", testName, err)
			}
		})
	}
}

// APIVerifier verifies that specific functions and types exist
type APIVerifier struct {
	name     string
	verifier func(t *testing.T) error
}

// NewAPIVerifier creates a new API verifier
func NewAPIVerifier(name string, verifier func(t *testing.T) error) *APIVerifier {
	return &APIVerifier{
		name:     name,
		verifier: verifier,
	}
}

// Name returns the name of the verifier
func (verifier *APIVerifier) Name() string {
	return verifier.name
}

// Setup is a no-op for API verifiers
func (verifier *APIVerifier) Setup(ctx context.Context) error {
	return nil
}

// Run runs the API verifier
func (verifier *APIVerifier) Run(ctx context.Context, version string, t *testing.T) error {
	return verifier.verifier(t)
}

// Teardown is a no-op for API verifiers
func (verifier *APIVerifier) Teardown(ctx context.Context) error {
	return nil
}

// LoadFixture loads a test fixture from the fixtures directory
func LoadFixture(t *testing.T, filename string) []byte {
	fixturesDir := filepath.Join("fixtures", filepath.Dir(filename))
	fixtureFile := filepath.Join(fixturesDir, filepath.Base(filename))

	data, err := os.ReadFile(fixtureFile)
	if err != nil {
		t.Fatalf("Failed to load fixture %s: %v", fixtureFile, err)
	}

	return data
}

// SaveFixture saves a test fixture to the fixtures directory
func SaveFixture(t *testing.T, filename string, data []byte) {
	fixturesDir := filepath.Join("fixtures", filepath.Dir(filename))
	fixtureFile := filepath.Join(fixturesDir, filepath.Base(filename))

	// Create directory if it doesn't exist
	if err := os.MkdirAll(fixturesDir, 0755); err != nil {
		t.Fatalf("Failed to create fixtures directory: %v", err)
	}

	// Write fixture
	if err := os.WriteFile(fixtureFile, data, 0644); err != nil {
		t.Fatalf("Failed to save fixture %s: %v", fixtureFile, err)
	}
}

// CompareVersions compares two version strings
func CompareVersions(a, b string) (int, error) {
	// Remove v prefix if present
	a = strings.TrimPrefix(a, "v")
	b = strings.TrimPrefix(b, "v")

	// Split version strings into parts
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	// Compare each part
	for i := 0; i < len(aParts) && i < len(bParts); i++ {
		// Try to parse as integer
		var aVal, bVal int
		if _, err := fmt.Sscanf(aParts[i], "%d", &aVal); err != nil {
			return 0, errors.New("invalid version format")
		}
		if _, err := fmt.Sscanf(bParts[i], "%d", &bVal); err != nil {
			return 0, errors.New("invalid version format")
		}

		// Compare
		if aVal < bVal {
			return -1, nil
		}
		if aVal > bVal {
			return 1, nil
		}
	}

	// If all parts are equal, compare lengths
	if len(aParts) < len(bParts) {
		return -1, nil
	}
	if len(aParts) > len(bParts) {
		return 1, nil
	}

	// Versions are equal
	return 0, nil
}

// VersionAtLeast checks if a version is at least the given minimum
func VersionAtLeast(version, minVersion string) (bool, error) {
	if version == "current" {
		// Current version is always considered to be at least any minimum
		return true, nil
	}

	// Compare versions
	result, err := CompareVersions(version, minVersion)
	if err != nil {
		return false, err
	}

	return result >= 0, nil
}

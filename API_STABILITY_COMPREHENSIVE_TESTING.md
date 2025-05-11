# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# Comprehensive Compatibility Testing

This document describes the comprehensive compatibility testing framework implemented as part of Phase 2 of the API Stability Implementation Plan for the Globus Go SDK.

## Overview

Comprehensive compatibility testing ensures that the SDK maintains backward compatibility across versions. This helps:

1. Detect breaking changes before release
2. Verify that integrations continue to work
3. Document compatibility guarantees
4. Support semantic versioning decisions

## Testing Framework

The compatibility testing framework is located in `tests/compatibility/` and consists of:

1. **Test Runner** - Executes compatibility tests
2. **Test Cases** - Verify specific aspects of compatibility
3. **Test Fixtures** - Reference data for comparison
4. **Utility Functions** - Version comparison, fixture loading, etc.

## Running Tests

The compatibility tests can be run using:

```bash
# Run against the current version
go test -v ./tests/compatibility/...

# Run against a specific version
VERSION=v0.9.15 go test -v ./tests/compatibility/...
```

For comprehensive testing across versions:

```bash
# Compare current with latest release
./scripts/run_compatibility_tests.sh

# Compare with a specific version
./scripts/run_compatibility_tests.sh --version v0.9.10
```

## Test Categories

The compatibility framework includes several categories of tests:

### 1. API Surface Tests

These tests verify that the public API surface remains compatible:

- **Required Methods** - Verify that required methods exist and have compatible signatures
- **Type Definitions** - Verify that types have the expected fields and methods
- **Constants and Variables** - Verify that constants and variables have expected values
- **Interface Implementations** - Verify that types implement required interfaces

Example:
```go
// Verify required methods
clientType := reflect.TypeOf(client)
method, found := clientType.MethodByName("GetTokenInfo")
if !found {
    t.Errorf("Required method GetTokenInfo not found")
}
```

### 2. Behavioral Tests

These tests verify that functions behave as expected:

- **Return Values** - Verify that functions return expected values
- **Error Handling** - Verify that functions handle errors properly
- **Context Handling** - Verify that functions respect context cancellation
- **Option Handling** - Verify that functional options work as expected

Example:
```go
// Verify behavior
result, err := client.GetTokenInfo(ctx, "test-token")
if err != nil {
    if errors.Is(err, auth.ErrInvalidToken) {
        // Expected error
    } else {
        t.Errorf("Unexpected error: %v", err)
    }
}
```

### 3. Integration Tests

These tests verify interactions between components:

- **Service Initialization** - Verify that service clients can be created
- **Authentication Flow** - Verify that authentication flows work
- **Data Transfer** - Verify that data transfer operations work
- **Error Propagation** - Verify that errors propagate correctly

Example:
```go
// Verify integration
authClient, _ := auth.NewClient()
tokenManager := auth.NewTokenManager(auth.TokenManagerOptions{})
err := tokenManager.AddToken(ctx, authClient, "test-token")
if err != nil {
    t.Errorf("Failed to add token: %v", err)
}
```

### 4. Compatibility with Dependent Projects

These tests verify that changes don't break dependent projects:

- **Downstream Tests** - Run tests for dependent projects
- **Sample Applications** - Compile and run sample applications
- **API Usage Patterns** - Verify common usage patterns work

Example:
```go
// Verify downstream project
cmd := exec.Command("go", "test", "-v", "./test-downstream-project/...")
output, err := cmd.CombinedOutput()
if err != nil {
    t.Errorf("Downstream project tests failed: %v\n%s", err, output)
}
```

## Continuous Integration

The comprehensive compatibility testing is integrated into the CI/CD pipeline:

1. **API Stability Workflow** - Runs on PRs and pushes to main
2. **Manual Trigger** - Can be run manually with specific versions
3. **Release Checks** - Runs before each release

### GitHub Actions Integration

The `api-stability.yml` workflow includes a job for comprehensive compatibility testing:

```yaml
compatibility-testing:
  name: Compatibility Testing
  runs-on: ubuntu-latest
  needs: [api-compatibility]
  steps:
    # ... setup steps ...
    - name: Run comprehensive compatibility tests
      run: ./scripts/run_compatibility_tests.sh
```

## Adding New Tests

To add a new compatibility test:

1. Create a new test file in the appropriate directory
2. Implement the `compatibility.Test` interface
3. Register the test in the init function
4. Add fixtures for comparison if needed

Example:
```go
// tests/compatibility/testcases/flows/client_test.go
package flows

import (
    "context"
    "testing"

    "github.com/scttfrdmn/globus-go-sdk/pkg/services/flows"
    "github.com/scttfrdmn/globus-go-sdk/tests/compatibility"
)

// FlowsClientTest verifies Flows client compatibility
type FlowsClientTest struct{}

func (t *FlowsClientTest) Name() string {
    return "FlowsClient"
}

func (t *FlowsClientTest) Setup(ctx context.Context) error {
    return nil
}

func (t *FlowsClientTest) Run(ctx context.Context, version string, t *testing.T) error {
    // Implement test
    return nil
}

func (t *FlowsClientTest) Teardown(ctx context.Context) error {
    return nil
}

func init() {
    compatibility.RegisterTest(&FlowsClientTest{})
}
```

## Best Practices

1. **Focus on Public API** - Test only the public API, not implementation details
2. **Version-Specific Tests** - Use `VersionAtLeast` to run tests only for applicable versions
3. **Clear Expectations** - Document what each test expects for compatibility
4. **Isolated Tests** - Make tests independent and self-contained
5. **Minimal Dependencies** - Minimize dependencies on external services
6. **Comprehensive Coverage** - Test all services and common usage patterns
7. **Fast Tests** - Keep tests fast to enable frequent running
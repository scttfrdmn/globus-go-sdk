# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# Compatibility Testing Framework

This directory contains the compatibility testing framework for the Globus Go SDK. The purpose of this framework is to systematically verify that changes to the SDK maintain backward compatibility with dependent code.

## Overview

The compatibility testing framework consists of:

1. Sample applications that use the SDK in various ways
2. Test fixtures that capture expected behaviors
3. Test runners that verify compatibility

## Test Structure

Each compatibility test follows this structure:

```
tests/compatibility/
├── testcases/
│   ├── auth/           # Auth service compatibility tests
│   ├── compute/        # Compute service compatibility tests
│   ├── flows/          # Flows service compatibility tests
│   ├── groups/         # Groups service compatibility tests
│   └── transfer/       # Transfer service compatibility tests
├── fixtures/           # Test fixtures and expected outputs
├── compat_runner.go    # Test runner
└── compat_test.go      # Main test file
```

## Running Tests

To run the compatibility tests:

```bash
go test -v ./tests/compatibility/...
```

To test compatibility with a specific version:

```bash
VERSION=v0.9.15 go test -v ./tests/compatibility/...
```

## Test Cases

Each test case verifies specific aspects of the SDK:

1. **API Surface Tests** - Verify that required functions and types exist
2. **Behavior Tests** - Verify that functions behave as expected
3. **Integration Tests** - Verify interactions between components
4. **Error Handling Tests** - Verify error handling semantics

## Adding New Tests

To add a new compatibility test:

1. Create a new test file in the appropriate service directory
2. Implement the test case using the `compat.Test` interface
3. Register the test in `compat_test.go`

Example:

```go
// tests/compatibility/testcases/auth/token_manager_test.go
package auth

import (
	"context"
	"testing"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/tests/compatibility"
)

// TokenManagerTest verifies token manager compatibility
type TokenManagerTest struct{}

func (t *TokenManagerTest) Name() string {
	return "TokenManager"
}

func (t *TokenManagerTest) Setup(ctx context.Context) error {
	return nil
}

func (t *TokenManagerTest) Run(ctx context.Context, version string, t *testing.T) error {
	// Create token manager
	tokenManager := auth.NewTokenManager(auth.TokenManagerOptions{})
	
	// Verify token manager methods exist and behave as expected
	if err := tokenManager.LoadTokens(ctx); err != auth.ErrNoTokensFound {
		t.Errorf("Expected ErrNoTokensFound, got %v", err)
	}
	
	return nil
}

func (t *TokenManagerTest) Teardown(ctx context.Context) error {
	return nil
}

func init() {
	compatibility.RegisterTest(&TokenManagerTest{})
}
```

## Best Practices

When adding compatibility tests:

1. Focus on behavior that dependent code relies on
2. Test both common and edge cases
3. Verify error conditions
4. Document expectations clearly
5. Keep tests isolated and repeatable
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# Testing Guide

_Last Updated: April 27, 2025_

This guide covers testing aspects of the Globus Go SDK.

> **DISCLAIMER**: The Globus Go SDK is an independent, community-developed project and is not officially affiliated with, endorsed by, or supported by Globus, the University of Chicago, or their affiliated organizations. Testing procedures and recommendations in this document are provided as guidance only.

## Table of Contents

- [Overview](#overview)
  - [Testing Approaches](#testing-approaches)
  - [Test Coverage Goals](#test-coverage-goals)
- [Unit Testing](#unit-testing)
  - [Structure](#unit-test-structure)
  - [Best Practices](#unit-testing-best-practices)
  - [Running Unit Tests](#running-unit-tests)
  - [Coverage Reports](#coverage-reports)
- [Integration Testing](#integration-testing)
  - [Prerequisites](#integration-testing-prerequisites)
  - [Setting Up Credentials](#setting-up-credentials)
  - [Creating a Globus App Registration](#creating-a-globus-app-registration)
  - [Setting Up Test Endpoints](#setting-up-test-endpoints)
  - [Writing Integration Tests](#writing-integration-tests)
  - [Running Integration Tests](#running-integration-tests)
  - [Test Data Safety](#test-data-safety)
- [Shell Script Testing](#shell-script-testing)
  - [ShellCheck Static Analysis](#shellcheck-static-analysis)
  - [BATS (Bash Automated Testing System)](#bats-bash-automated-testing-system)
  - [Writing BATS Tests](#writing-bats-tests)
  - [Running Shell Tests](#running-shell-tests)
  - [Shell Testing Best Practices](#shell-testing-best-practices)
- [Security Testing](#security-testing)
  - [Static Analysis](#static-analysis)
  - [Dependency Scanning](#dependency-scanning)
  - [Secret Detection](#secret-detection)
  - [Token Analysis](#token-analysis)
  - [Security Test Interpretation](#security-test-interpretation)
  - [Addressing Security Issues](#addressing-security-issues)
- [CI/CD Testing](#cicd-testing)
  - [GitHub Actions Workflows](#github-actions-workflows)
  - [Pipeline Structure](#pipeline-structure)
  - [Test Environment Variables](#test-environment-variables)
  - [Running Tests in CI](#running-tests-in-ci)
- [Troubleshooting](#troubleshooting)
  - [Common Issues](#common-issues)
  - [Debugging Test Failures](#debugging-test-failures)
  - [Test Environment Issues](#test-environment-issues)
- [Resources](#resources)
  - [Tools](#tools)
  - [Documentation](#documentation)

## Overview

The Globus Go SDK employs a comprehensive testing strategy to ensure code quality, functionality, and security. Our testing approach includes unit tests, integration tests, shell script tests, and security tests.

### Testing Approaches

The SDK uses the following testing approaches:

1. **Unit Testing**: Tests individual components in isolation
2. **Integration Testing**: Tests interactions with real Globus services
3. **API Export Verification**: Ensures that all required functions and interfaces are properly exported
4. **Shell Script Testing**: Validates shell scripts using static analysis and automated tests
5. **Security Testing**: Identifies potential security vulnerabilities

### Test Coverage Goals

- Aim for >80% code coverage
- All exported functions must have tests
- Error paths should be tested
- Mock external dependencies for unit testing

## Unit Testing

Unit tests validate individual components of the SDK in isolation, typically by mocking external dependencies.

### Unit Test Structure

- Located in `*_test.go` files in the same package as the code they test
- Follow Go's standard testing conventions
- Use table-driven tests for comprehensive test cases
- Mock external dependencies for isolation

Example unit test structure:

```go
func TestFunction(t *testing.T) {
    // Test setup
    client := NewClient()
    mockServer := setupMockServer()
    defer mockServer.Close()
    
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"happy path", "valid-input", "expected-output", false},
        {"error case", "invalid-input", "", true},
    }
    
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            result, err := client.Function(tc.input)
            
            if tc.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tc.expected, result)
            }
        })
    }
}
```

### Unit Testing Best Practices

1. **Use table-driven tests**: Test multiple cases within a single test function
2. **Test both success and error paths**: Ensure error handling works correctly
3. **Mock external dependencies**: Use httptest package for HTTP dependencies
4. **Use subtests**: Organize tests with `t.Run()` for better reporting
5. **Keep tests focused**: Test one specific behavior per test
6. **Use assertion libraries**: Consider testify/assert for cleaner assertions
7. **Clean up resources**: Use `defer` to clean up test resources

### Running Unit Tests

Run all unit tests:

```bash
go test ./...
```

Run tests for a specific package:

```bash
go test ./pkg/services/auth/...
```

Run a specific test:

```bash
go test -run TestSpecificFunction ./pkg/services/auth
```

Run tests with verbose output:

```bash
go test -v ./...
```

### Coverage Reports

Generate test coverage:

```bash
# Run tests with coverage
go test -cover ./...

# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out
```

Using make targets:

```bash
# Run tests with coverage reporting
make test-coverage
```

## Integration Testing

Integration tests verify that the SDK correctly interacts with real Globus services.

### Integration Testing Prerequisites

To run integration tests, you need:

1. A Globus account with appropriate permissions
2. A Globus client application with necessary scopes
3. Access to at least one Globus endpoint for transfer tests
4. Go installed on your system (version 1.18 or later)

### Setting Up Credentials

Integration tests require real Globus API credentials provided as environment variables:

#### Required Variables

- `GLOBUS_TEST_CLIENT_ID`: Your Globus client ID
- `GLOBUS_TEST_CLIENT_SECRET`: Your Globus client secret

#### Optional Variables (Recommended for Transfer Tests)

- `GLOBUS_TEST_SOURCE_ENDPOINT_ID`: ID of a source endpoint for transfer tests
- `GLOBUS_TEST_DESTINATION_ENDPOINT_ID`: ID of a destination endpoint for transfer tests
- `GLOBUS_TEST_SOURCE_PATH`: Path on the source endpoint (default: `/globus-test`)
- `GLOBUS_TEST_DESTINATION_PATH`: Path on the destination endpoint (default: `/globus-test`)

#### Optional Variables (For Group Tests)

- `GLOBUS_TEST_GROUP_ID`: ID of a Globus group for testing group operations
- `GLOBUS_TEST_USER_ID`: ID of a user for testing group membership operations

#### Using an Environment File

You can create a `.env.test` file in the project root:

```
GLOBUS_TEST_CLIENT_ID=your-client-id
GLOBUS_TEST_CLIENT_SECRET=your-client-secret
GLOBUS_TEST_SOURCE_ENDPOINT_ID=your-source-endpoint-id
GLOBUS_TEST_DESTINATION_ENDPOINT_ID=your-destination-endpoint-id
GLOBUS_TEST_SOURCE_PATH=/globus-test
GLOBUS_TEST_DESTINATION_PATH=/globus-test
GLOBUS_TEST_USER_ID=your-user-id
GLOBUS_TEST_GROUP_ID=your-group-id
```

⚠️ **IMPORTANT: Never commit this file to the repository.** The `.env.test` file is included in `.gitignore` to prevent accidental exposure of credentials.

### Creating a Globus App Registration

To access Globus APIs, create an app registration:

1. Go to [https://developers.globus.org/](https://developers.globus.org/)
2. Log in with your Globus account
3. Click "Register your app with Globus"
4. Fill in the required fields:
   - App Name: "Globus Go SDK Testing"
   - Contact Email: Your email
   - Redirect URLs: `https://localhost:8000/callback` (for testing)
   - Scopes: Select all applicable scopes (at minimum: openid, profile, email, urn:globus:auth:scope:transfer.api.globus.org:all, urn:globus:auth:scope:groups.api.globus.org:all)
5. Click "Create App"
6. Note your Client ID and Client Secret (needed for tests)

### Setting Up Test Endpoints

For transfer tests, you need access to at least two endpoints:

#### Option A: Use Personal Endpoints (Recommended for Initial Testing)

1. **Create a Personal Endpoint**:
   - Install Globus Connect Personal from [https://www.globus.org/globus-connect-personal](https://www.globus.org/globus-connect-personal)
   - Follow the setup instructions to create a personal endpoint
   - Note your endpoint ID

2. **Set up Test Directories**:
   - Create a directory named `/globus-test` on your personal endpoint
   - Create a few test files in this directory

3. **Use the Same Endpoint for Both Source and Destination**:
   - For initial testing, you can use the same personal endpoint as both source and destination
   - Just use different subdirectories for source and destination paths

#### Option B: Use Existing Endpoints

If you have access to existing Globus endpoints:

1. **Identify Two Endpoints**:
   - Choose one endpoint as source and one as destination
   - Ensure you have write access to both endpoints
   - Note both endpoint IDs

2. **Create Test Directories**:
   - Create `/globus-test` directories on both endpoints
   - Add some test files to the source endpoint directory

### Writing Integration Tests

When writing integration tests:

1. Use the `_test` suffix in the test file name
2. Use the `//go:build integration` build tag at the top of the file
3. Name test functions with the `TestIntegration_` prefix
4. Use the `getTestCredentials` function to get credentials
5. Include cleanup code to remove any created resources
6. Make tests skip gracefully if required credentials are missing

Example structure:

```go
//go:build integration
package mypackage_test

import (
    "testing"
    "context"
    "os"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/myservice"
)

func TestIntegration_MyFeature(t *testing.T) {
    // Get credentials
    clientID := os.Getenv("GLOBUS_TEST_CLIENT_ID")
    clientSecret := os.Getenv("GLOBUS_TEST_CLIENT_SECRET")
    
    if clientID == "" || clientSecret == "" {
        t.Skip("Integration test requires GLOBUS_TEST_CLIENT_ID and GLOBUS_TEST_CLIENT_SECRET")
    }
    
    // Create client and run test
    // ...
    
    // Clean up resources
    defer func() {
        // Delete any created resources
    }()
}
```

### Running Integration Tests

#### Using the Script

The easiest way to run integration tests is using the provided script:

```bash
# Run all integration tests
./scripts/run_integration_tests.sh

# Run tests for a specific package
./scripts/run_integration_tests.sh pkg/services/transfer

# Run tests for a specific test function
./scripts/run_integration_tests.sh pkg/services/transfer TestIntegration_ResumableTransfer
```

The script automatically:
1. Loads environment variables from `.env.test` if present
2. Checks that required variables are set
3. Runs the specified tests with the appropriate tags

#### Manual Execution

Alternatively, run the tests manually:

```bash
# Set environment variables (if not using .env.test)
export GLOBUS_TEST_CLIENT_ID=your-client-id
export GLOBUS_TEST_CLIENT_SECRET=your-client-secret
# ... set other variables as needed

# Run all integration tests
go test -v -tags=integration ./...

# Run tests for a specific package
go test -v -tags=integration ./pkg/services/transfer/...

# Run a specific test
go test -v -tags=integration ./pkg/services/transfer -run TestIntegration_ResumableTransfer
```

### Test Data Safety

To avoid data loss or unintended consequences:

1. Always use test-specific paths and resources
2. Include timestamps in test resource names to avoid conflicts
3. Clean up all created resources when tests complete
4. Use read-only operations where possible
5. Never modify or delete data outside of your test directories

## API Export Verification

API Export Verification ensures that all functions, types, and interfaces required by dependent projects are properly exported from the SDK. This helps prevent breaking changes by catching issues with missing or mis-exported APIs.

### Export Verification Tools

The SDK includes dedicated tools for verifying API exports:

1. **`verify_package_exports.go`**: A standalone tool that verifies required exports
2. **`verify_exports.sh`**: A shell script wrapper for easy use in CI/CD pipelines
3. **HTTP Pool API tests**: Dedicated tests for HTTP connection pool exports

### Running Export Verification

```bash
# Run the export verification script
./scripts/verify_exports.sh

# Run HTTP pool API tests
go test -v ./pkg/core/http/pool_api_test.go

# Run connection pool integration tests
go test -v ./pkg/connection_pools_test.go
```

### Writing Export Verification Tests

When adding new exported functions or interfaces:

1. Add them to the list of critical exports in `scripts/verify_package_exports.go`
2. Add interface implementation tests for any new interfaces
3. Add practical usage tests that simulate how dependent packages use the API

Example:

```go
// Add to verify_package_exports.go
criticalExports := []requiredExport{
    // Existing exports...
    {"MyNewFunction", "github.com/scttfrdmn/globus-go-sdk/pkg/mypackage", mypackage.MyNewFunction, false},
}

// Add interface implementation test
func TestMyInterfaceImplementation(t *testing.T) {
    var _ MyInterface = (*MyConcreteType)(nil)  // Static check
    
    // Runtime check
    instance := NewMyConcreteType()
    _, ok := interface{}(instance).(MyInterface)
    if !ok {
        t.Error("MyConcreteType does not implement MyInterface")
    }
}
```

### Best Practices for API Export Verification

1. **Test direct usage patterns**: Verify the exact patterns that dependent projects use
2. **Use direct type assertions**: For interface implementation tests, use direct type assertions instead of reflect-based checks
3. **Check for nil functions**: Verify that exported functions are not nil
4. **Add tests for critical components**: Focus on APIs used by dependent projects
5. **Test practical usage scenarios**: Go beyond simple existence checks to verify that the APIs work as expected

## Shell Script Testing

The Globus Go SDK uses two main tools for ensuring shell script quality:

1. **ShellCheck**: A static analysis tool for shell scripts that provides warnings and suggestions for bash/sh shell scripts
2. **BATS (Bash Automated Testing System)**: A testing framework for Bash that provides a TAP-compliant testing experience

### ShellCheck Static Analysis

ShellCheck is used to lint shell scripts and identify potential issues.

#### Installing ShellCheck

- **macOS**: `brew install shellcheck`
- **Ubuntu/Debian**: `apt-get install shellcheck`
- **Other platforms**: See [ShellCheck's installation guide](https://github.com/koalaman/shellcheck#installing)

#### Running ShellCheck

```sh
# Lint all shell scripts
./scripts/lint_shell_scripts.sh

# Or use the Makefile target
make lint-shell
```

#### ShellCheck Configuration

ShellCheck is configured via `.shellcheckrc` in the project root, containing:

- Disabled checks that aren't relevant to our project
- Project-specific settings

### BATS (Bash Automated Testing System)

BATS is a testing framework for Bash scripts.

#### Installing BATS

The project includes a script to install BATS and its dependencies:

```sh
# Install BATS
./scripts/install_bats.sh
```

This installs:
- `bats-core`: The main BATS testing framework
- `bats-support`: Helper functions for better output
- `bats-assert`: Assertion functions for easier testing
- `bats-file`: File-related assertion functions

### Writing BATS Tests

BATS tests are written in `.bats` files located in the `tests/` directory:

```bash
#!/usr/bin/env bats
# Load helper libraries
load "bats/bats-support/load.bash"
load "bats/bats-assert/load.bash"
load "bats/bats-file/load.bash"

# Test function
@test "Example test" {
  run echo "Hello, world!"
  assert_success
  assert_output "Hello, world!"
}
```

Example test for a script that counts files in a directory:

```bash
# Script: scripts/count_files.sh
#!/bin/bash
# Count files in a directory
count_files() {
  local dir="$1"
  find "$dir" -type f | wc -l
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  count_files "$@"
fi
```

```bash
# Test: tests/test_count_files.bats
#!/usr/bin/env bats
load "bats/bats-support/load.bash"
load "bats/bats-assert/load.bash"

# Source the script to test
source "../scripts/count_files.sh"

setup() {
  # Create a temporary directory
  TEST_DIR="$(mktemp -d)"
  
  # Create some files
  touch "$TEST_DIR/file1.txt"
  touch "$TEST_DIR/file2.txt"
}

teardown() {
  # Clean up
  rm -rf "$TEST_DIR"
}

@test "count_files counts files correctly" {
  # Run the function
  run count_files "$TEST_DIR"
  
  # Assert the output
  assert_output "2"
}
```

### Running Shell Tests

```sh
# Run all BATS tests
./scripts/run_shell_tests.sh

# Or use the Makefile target
make test-shell

# Run a specific test file
./scripts/run_shell_tests.sh tests/test_specific_script.bats
```

### Shell Testing Best Practices

1. **Always run ShellCheck**: Before committing any shell script changes, run ShellCheck
2. **Write tests**: Add BATS tests for any new shell script functionality
3. **Keep scripts modular**: Make functions testable by keeping them small and focused
4. **Use `set -e`**: Include `set -e` in scripts to fail fast on errors
5. **Document scripts**: Add comments explaining the purpose and usage of scripts and functions
6. **Setup and teardown**: Use `setup()` and `teardown()` functions for test preparation and cleanup
7. **Mocking**: Create mock functions to simulate commands and isolate unit tests
8. **Temporary directories**: Use temporary directories for test file operations
9. **Helper functions**: Extract common test logic into helper functions
10. **Assertions**: Use the assertion libraries (bats-assert, bats-file) for clearer tests

## Security Testing

The Globus Go SDK implements several layers of security testing to ensure the safety and integrity of the codebase.

### Static Analysis

Static analysis tools scan code for potential security issues without executing it.

#### Setting Up Static Analysis

```bash
# Install gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest
```

#### Running Static Analysis

```bash
# Run gosec on the entire codebase
gosec ./...

# Run gosec with JSON output
gosec -fmt=json -out=gosec-results.json ./...

# Focus on high-severity issues only
gosec -severity=high ./...
```

### Dependency Scanning

Dependency scanning identifies vulnerabilities in third-party packages used by the SDK.

#### Setting Up Dependency Scanning

```bash
# Install nancy
go install github.com/sonatype-nexus-community/nancy@latest
```

#### Running Dependency Scanning

```bash
# Scan all dependencies
go list -json -m all | nancy sleuth

# Exclude development dependencies
go list -json -m all | nancy sleuth --exclude-dev

# Output as JSON
go list -json -m all | nancy sleuth --output json > nancy-results.json
```

### Secret Detection

Secret detection tools identify potential secrets and credentials in the codebase.

#### Setting Up Secret Detection

```bash
# Install gitleaks
go install github.com/zricethezav/gitleaks/v8@latest
```

#### Running Secret Detection

```bash
# Scan current directory
gitleaks detect

# Scan with custom configuration
gitleaks detect --config gitleaks.toml

# Output as JSON
gitleaks detect --report-format json --report-path gitleaks-report.json
```

### Token Analysis

The SDK provides a dedicated security test tool for token analysis:

```bash
# Build the tool
go build -o security-test ./cmd/security-test

# Run security self-test
./security-test -self

# Analyze token for security issues
./security-test -token "your_token" -client-id "your_client_id"
```

### Security Test Interpretation

#### gosec Results

gosec categorizes findings by severity and rule ID. Key rules include:

- **G101** - Hardcoded credentials
- **G102** - Binding to all network interfaces
- **G104** - Unhandled errors
- **G107** - URL provided to HTTP request as taint input
- **G402** - TLS InsecureSkipVerify set true
- **G404** - Weak random number generator (math/rand instead of crypto/rand)

#### nancy Results

nancy provides information about vulnerabilities in dependencies:

- **CVE ID** - The Common Vulnerabilities and Exposures identifier
- **CVSS Score** - The Common Vulnerability Scoring System score (0-10)
- **Affected Package** - The affected dependency
- **Vulnerable Versions** - The affected versions
- **Fixed Version** - The version where the vulnerability is fixed

#### gitleaks Results

gitleaks detects potential secrets in the codebase:

- **Rule** - The rule that triggered the detection
- **Secret** - A masked version of the detected secret
- **File** - The file where the secret was found
- **Line** - The line number where the secret was found

### Addressing Security Issues

#### Priority Levels

1. **Critical** - Must be fixed immediately
   - Authentication/authorization bypasses
   - Remote code execution
   - Token leakage
   - High CVSS (9.0-10.0) vulnerabilities

2. **High** - Must be fixed in the next release
   - Information disclosure
   - Medium-High CVSS (7.0-8.9) vulnerabilities
   - Sensitive data exposure

3. **Medium** - Should be fixed in a timely manner
   - Low-Medium CVSS (4.0-6.9) vulnerabilities
   - Insecure configurations

4. **Low** - Fix when convenient
   - Low CVSS (0.1-3.9) vulnerabilities
   - Code quality issues

#### Remediation Process

1. **Triage**: Assess the severity and impact
2. **Document**: Create an issue with details about the vulnerability
3. **Test**: Create a test case that reproduces the issue
4. **Fix**: Implement a fix
5. **Verify**: Ensure the fix resolves the issue
6. **Release**: Include the fix in the next appropriate release

#### False Positives

If you identify a false positive:

1. Document the finding and why it's a false positive
2. Add an appropriate comment to the code:
   ```go
   // gosec:ignore:G404 Using math/rand is acceptable for non-cryptographic purposes
   ```
3. Configure the tool to exclude the false positive in future scans

## Local Testing with Git Hooks

Git hooks provide an effective way to run tests locally before code is committed or pushed to the repository.

### Available Git Hooks

The Globus Go SDK provides the following Git hooks:

1. **Pre-commit Hook**:
   - Runs license header checks
   - Formats code with `go fmt`
   - Runs linting with `staticcheck` (if installed)
   - Performs static analysis with `go vet`
   - Runs unit tests in short mode

2. **Pre-push Hook**:
   - Runs all tests (including integration tests)
   - Checks documentation
   - Performs security scanning

### Installing Git Hooks

To install all hooks at once:

```bash
./scripts/install-all-hooks.sh
```

To install specific hooks:

```bash
# Install only pre-commit hook
./scripts/install-hooks.sh

# Install only pre-push hook
./scripts/install-pre-push-hook.sh
```

For more details, see [Git Hooks](git-hooks.md).

## CI/CD Testing

The Globus Go SDK includes comprehensive CI/CD pipelines to automatically test code changes.

### GitHub Actions Workflows

The repository includes several GitHub Actions workflows:

1. **Main workflow** (`go.yml`):
   - Runs on each PR and push to main
   - Performs linting, testing, and building
   - Checks multiple Go versions

2. **Security Scan** (`security-scan.yml`):
   - Dedicated workflow for comprehensive security scanning
   - Runs gosec, nancy, and gitleaks

3. **Shell Lint** (`shell-lint.yml`):
   - Workflow for shell script linting with shellcheck

4. **Integration Tests** (`integration-tests.yml`):
   - Runs on manual trigger and schedule
   - Executes integration tests against real Globus services

### Pipeline Structure

The standard CI pipeline follows this structure:

1. **Setup**: Prepare the environment with the correct Go version
2. **Lint**: Run golangci-lint to verify code style
3. **Test**: Run unit tests with coverage reporting
4. **Build**: Verify the project builds successfully
5. **API Export Verification**: Run API availability tests and verify package exports
6. **Security Scan**: Run security tools (gosec, nancy, gitleaks)
7. **Report**: Upload test coverage and other reports

### Test Environment Variables

For CI/CD environments, store credentials as encrypted secrets:

```yaml
env:
  GLOBUS_TEST_CLIENT_ID: ${{ secrets.GLOBUS_TEST_CLIENT_ID }}
  GLOBUS_TEST_CLIENT_SECRET: ${{ secrets.GLOBUS_TEST_CLIENT_SECRET }}
  GLOBUS_TEST_SOURCE_ENDPOINT_ID: ${{ secrets.GLOBUS_TEST_SOURCE_ENDPOINT_ID }}
  GLOBUS_TEST_DEST_ENDPOINT_ID: ${{ secrets.GLOBUS_TEST_DEST_ENDPOINT_ID }}
  GLOBUS_TEST_USER_ID: ${{ secrets.GLOBUS_TEST_USER_ID }}
  GLOBUS_TEST_GROUP_ID: ${{ secrets.GLOBUS_TEST_GROUP_ID }}
```

### Running Tests in CI

Example GitHub Actions configurations:

#### Integration Tests Workflow

```yaml
name: Integration Tests

on:
  workflow_dispatch:  # Manual trigger only
  schedule:
    - cron: '0 0 * * 0'  # Weekly on Sundays

jobs:
  integration-tests:
    runs-on: ubuntu-latest
    
    env:
      GLOBUS_TEST_CLIENT_ID: ${{ secrets.GLOBUS_TEST_CLIENT_ID }}
      GLOBUS_TEST_CLIENT_SECRET: ${{ secrets.GLOBUS_TEST_CLIENT_SECRET }}
      GLOBUS_TEST_SOURCE_ENDPOINT_ID: ${{ secrets.GLOBUS_TEST_SOURCE_ENDPOINT_ID }}
      GLOBUS_TEST_DEST_ENDPOINT_ID: ${{ secrets.GLOBUS_TEST_DEST_ENDPOINT_ID }}
      GLOBUS_TEST_USER_ID: ${{ secrets.GLOBUS_TEST_USER_ID }}
      GLOBUS_TEST_GROUP_ID: ${{ secrets.GLOBUS_TEST_GROUP_ID }}
    
    steps:
      - uses: actions/checkout@v2
      
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: '1.21'
          
      - name: Run integration tests
        run: ./scripts/run_integration_tests.sh
```

#### API Export Verification Workflow

```yaml
name: API Export Verification

on:
  push:
    branches: [ main, v0.9.0-release ]
  pull_request:
    branches: [ main, v0.9.0-release ]

jobs:
  verify-exports:
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/checkout@v2
      
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: '1.21'
          
      - name: Get dependencies
        run: |
          go mod download
          go mod verify
          
      - name: Verify package exports
        run: ./scripts/verify_exports.sh
        
      - name: Run HTTP pool API tests
        run: go test -v ./pkg/core/http/...
        
      - name: Run connection pool integration tests
        run: go test -v ./pkg/connection_pools_test.go
```

Best practices for CI/CD testing:

1. Use a dedicated test account with limited permissions
2. Ensure tests are idempotent and clean up after themselves
3. Store credentials securely as encrypted secrets
4. Run integration tests on a schedule rather than every commit
5. Implement timeouts to avoid hung tests

## Troubleshooting

### Common Issues

#### Invalid Credentials

**Symptoms**: Authentication errors, "invalid_client" error messages

**Solutions**:
- Double-check client ID and secret for typos
- Ensure the app is still active on the Developers Dashboard
- Create a new client secret if needed

#### Endpoint Access Issues

**Symptoms**: "Permission denied" errors, endpoint not found

**Solutions**:
- Verify endpoint IDs are correct
- Ensure endpoints are activated
- Check path permissions on the endpoints
- Make sure test directories exist

#### Rate Limiting

**Symptoms**: HTTP 429 errors, "too many requests" messages

**Solutions**:
- Add delays between tests
- Reduce concurrent test execution
- Implement exponential backoff in tests

### Debugging Test Failures

- Run failed tests with verbose output:
  ```bash
  go test -v -run TestSpecificTest ./path/to/package
  ```

- Use environment variable for debug logging:
  ```bash
  GLOBUS_SDK_LOG_LEVEL=debug go test ./...
  ```

- Isolate the failing test:
  ```bash
  # Run just the failing test
  go test -v -run TestSpecific ./path/to/package
  
  # Add print statements or logging
  // Add this to the test
  t.Logf("Value: %v", value)
  ```

### Test Environment Issues

- **Go version conflicts**: Ensure your Go version matches the project requirements
- **Missing dependencies**: Run `go mod tidy` to resolve dependency issues
- **Permission issues**: Check file and directory permissions
- **Network connectivity**: Ensure you have internet access for integration tests

## Resources

### Tools

- [Go Testing Package](https://golang.org/pkg/testing/)
- [ShellCheck](https://github.com/koalaman/shellcheck)
- [BATS](https://github.com/bats-core/bats-core)
- [gosec](https://github.com/securego/gosec)
- [nancy](https://github.com/sonatype-nexus-community/nancy)
- [gitleaks](https://github.com/zricethezav/gitleaks)

### Documentation

- [Globus API Documentation](https://docs.globus.org/api/)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go)
- [OWASP Go Security Cheatsheet](https://github.com/OWASP/CheatSheetSeries/blob/master/cheatsheets/Go_Security_Cheatsheet.md)
- [Shell Script Best Practices](https://kvz.io/bash-best-practices.html)
- [HTTP Pool API Testing](../../HTTP_POOL_API_TESTING.md) - Guide for HTTP pool API verification tests
- [Testing Goals](test-goals.md) - Testing goals and priorities for the SDK
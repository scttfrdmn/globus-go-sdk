#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# Comprehensive testing script for the Globus Go SDK
# This script runs all tests, examples, and validation with real Globus credentials

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SDK_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
ENV_FILE="${SDK_ROOT}/.env.test"
LOG_FILE="${SDK_ROOT}/comprehensive_testing.log"

# Function to log messages
log() {
    local message="$1"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $message" | tee -a "${LOG_FILE}"
}

# Function to run tests and log results
run_test() {
    local test_name="$1"
    local command="$2"
    local description="$3"
    
    log "========================================"
    log "STARTING: ${test_name}"
    log "DESCRIPTION: ${description}"
    log "COMMAND: ${command}"
    log "========================================"
    
    if eval "${command}"; then
        log "RESULT: PASS - ${test_name}"
        return 0
    else
        log "RESULT: FAIL - ${test_name}"
        return 1
    fi
}

# Initialize log file
> "${LOG_FILE}"
log "Starting comprehensive testing for Globus Go SDK"
log "SDK root: ${SDK_ROOT}"

# Check for .env.test file
if [ ! -f "${ENV_FILE}" ]; then
    log "ERROR: .env.test file not found in SDK root directory."
    log "Please create an .env.test file with your Globus credentials."
    exit 1
fi

# Source the environment variables
source "${ENV_FILE}"

# Check for required environment variables
# Support both prefixed and non-prefixed variables
if [ -n "${GLOBUS_TEST_CLIENT_ID}" ]; then
    export GLOBUS_CLIENT_ID="${GLOBUS_TEST_CLIENT_ID}"
fi
if [ -n "${GLOBUS_TEST_CLIENT_SECRET}" ]; then
    export GLOBUS_CLIENT_SECRET="${GLOBUS_TEST_CLIENT_SECRET}"
fi
if [ -n "${GLOBUS_TEST_SOURCE_ENDPOINT_ID}" ]; then
    export SOURCE_ENDPOINT_ID="${GLOBUS_TEST_SOURCE_ENDPOINT_ID}"
fi
if [ -n "${GLOBUS_TEST_DEST_ENDPOINT_ID}" ]; then
    export DEST_ENDPOINT_ID="${GLOBUS_TEST_DEST_ENDPOINT_ID}"
fi

# Now check if we have what we need
if [ -z "${GLOBUS_CLIENT_ID}" ] || [ -z "${GLOBUS_CLIENT_SECRET}" ]; then
    log "ERROR: GLOBUS_CLIENT_ID and/or GLOBUS_CLIENT_SECRET missing from .env.test file."
    log "Make sure you have either GLOBUS_CLIENT_ID or GLOBUS_TEST_CLIENT_ID set."
    exit 1
fi

log "Found credentials in .env.test"
log "Testing with client ID: ${GLOBUS_CLIENT_ID}"

# 1. Run all unit tests
run_test "Unit Tests" "cd ${SDK_ROOT} && go test ./pkg/..." "Run all unit tests in the pkg directory"

# 2. Run all examples to ensure they compile
run_test "Example Compilation" "cd ${SDK_ROOT} && go build ./examples/..." "Ensure all examples compile"

# 3. Test token management with real credentials
run_test "Token Management" "cd ${SDK_ROOT}/examples/token-management && ./test_tokens.sh" "Test token management with real credentials"

# 4. Run verification credentials tool
run_test "Verify Credentials" "cd ${SDK_ROOT}/cmd/verify-credentials && go build && ./verify-credentials" "Verify credentials with the standalone tool"

# 5. Run integration tests for all services
run_test "Integration Tests" "${SDK_ROOT}/scripts/run_integration_tests.sh" "Run integration tests for all services"

# 6. Run specific examples with real credentials
log "Running specific examples with real credentials..."

# Auth example
run_test "Auth Example" "cd ${SDK_ROOT}/cmd/examples/auth && go run main.go" "Test auth example with real credentials"

# Transfer example
run_test "Transfer Example" "cd ${SDK_ROOT}/cmd/examples/transfer && go run main.go --list-endpoints" "Test transfer example with real credentials"

# Groups example
run_test "Groups Example" "cd ${SDK_ROOT}/cmd/examples/groups && go run main.go --list" "Test groups example with real credentials"

# Search example
run_test "Search Example" "cd ${SDK_ROOT}/cmd/examples/search && go run main.go --list-indexes" "Test search example with real credentials"

# Flows example
run_test "Flows Example" "cd ${SDK_ROOT}/cmd/examples/flows && go run main.go --list" "Test flows example with real credentials"

# 7. Run linting checks
run_test "Linting" "cd ${SDK_ROOT} && go vet ./..." "Run Go vet linting checks"

# 8. Security scan
if command -v gosec &> /dev/null; then
    run_test "Security Scan" "cd ${SDK_ROOT} && gosec ./..." "Run security scan with gosec"
else
    log "SKIPPED: Security scan - gosec not installed"
fi

# Summary
log "========================================"
log "TESTING COMPLETED"
log "See comprehensive_testing.log for detailed results"
log "========================================"

# Check for failures in the log
if grep -q "RESULT: FAIL" "${LOG_FILE}"; then
    log "OVERALL RESULT: SOME TESTS FAILED"
    grep "RESULT: FAIL" "${LOG_FILE}"
    exit 1
else
    log "OVERALL RESULT: ALL TESTS PASSED"
    exit 0
fi
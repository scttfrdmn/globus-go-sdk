#!/usr/bin/env bats
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2025 Scott Friedman and Project Contributors

# Load the BATS test helper libraries
load "bats/bats-support/load.bash"
load "bats/bats-assert/load.bash"
load "bats/bats-file/load.bash"

# Path to the script
SCRIPT_PATH="../scripts/run_integration_tests.sh"

# Setup function runs before each test
setup() {
    # Create a temporary directory for tests
    TEST_TEMP_DIR="$(mktemp -d)"
    
    # Export as environment variable for use in tests
    export TEST_TEMP_DIR
    
    # Create a mock .env.test file
    cat > "${TEST_TEMP_DIR}/.env.test" << EOF
GLOBUS_TEST_CLIENT_ID=test-client-id
GLOBUS_TEST_CLIENT_SECRET=test-client-secret
GLOBUS_TEST_SOURCE_ENDPOINT_ID=test-source-endpoint
GLOBUS_TEST_DEST_ENDPOINT_ID=test-dest-endpoint
GLOBUS_TEST_GROUP_ID=test-group-id
EOF
}

# Teardown function runs after each test
teardown() {
    # Remove the temporary directory
    rm -rf "$TEST_TEMP_DIR"
}

# Helper function to mock go test command
mock_go_test() {
    cat > "${TEST_TEMP_DIR}/mock_go_test.sh" << 'EOF'
#!/bin/bash
echo "Running mock go test with args: $@"
echo "MOCK_GO_TEST_CALLED=true" >> "${TEST_TEMP_DIR}/go_test_called"
echo "MOCK_GO_TEST_ARGS=$@" >> "${TEST_TEMP_DIR}/go_test_called"
EOF
    chmod +x "${TEST_TEMP_DIR}/mock_go_test.sh"
    
    # Add the directory with the mock to PATH
    export PATH="${TEST_TEMP_DIR}:$PATH"
    
    # Create a mock alias for 'go'
    alias go="${TEST_TEMP_DIR}/mock_go_test.sh"
}

# Test basic script execution
@test "Script runs without errors" {
    cd "$TEST_TEMP_DIR"
    run bash "$SCRIPT_PATH" -- echo "Test only" || exit 0
    
    assert_success
    assert_output --partial "Test only"
}

# Test env file loading
@test "Loads environment variables from .env.test file" {
    cd "$TEST_TEMP_DIR"
    
    # Run the function to load env vars (sourcing the script)
    run bash -c "source $SCRIPT_PATH && load_env_file && echo \$GLOBUS_TEST_CLIENT_ID"
    
    assert_success
    assert_output --partial "test-client-id"
}

# Test check_env_vars success
@test "check_env_vars succeeds when required vars are set" {
    cd "$TEST_TEMP_DIR"
    
    export GLOBUS_TEST_CLIENT_ID="test-client-id"
    export GLOBUS_TEST_CLIENT_SECRET="test-client-secret"
    
    run bash -c "source $SCRIPT_PATH && check_env_vars"
    
    assert_success
}

# Test check_env_vars failure
@test "check_env_vars fails when required vars are missing" {
    cd "$TEST_TEMP_DIR"
    
    # Unset any existing variables
    unset GLOBUS_TEST_CLIENT_ID
    unset GLOBUS_TEST_CLIENT_SECRET
    
    run bash -c "source $SCRIPT_PATH && check_env_vars"
    
    assert_failure
    assert_output --partial "Some required environment variables are missing"
}

# Test running all integration tests
@test "Runs all integration tests when no arguments provided" {
    cd "$TEST_TEMP_DIR"
    mock_go_test
    
    # Source the script and call main function to simulate execution
    run bash -c "source $SCRIPT_PATH && main"
    
    assert_success
    assert [ -f "${TEST_TEMP_DIR}/go_test_called" ]
    assert_file_contains "${TEST_TEMP_DIR}/go_test_called" "MOCK_GO_TEST_ARGS=-v -tags=integration ./..."
}

# Test running specific package tests
@test "Runs tests for specific package when package provided" {
    cd "$TEST_TEMP_DIR"
    mock_go_test
    
    # Source the script and call main function with a package argument
    run bash -c "source $SCRIPT_PATH && main pkg/services/auth"
    
    assert_success
    assert [ -f "${TEST_TEMP_DIR}/go_test_called" ]
    assert_file_contains "${TEST_TEMP_DIR}/go_test_called" "MOCK_GO_TEST_ARGS=-v -tags=integration ./pkg/services/auth/..."
}

# Test running specific tests with pattern
@test "Runs tests matching pattern when package and pattern provided" {
    cd "$TEST_TEMP_DIR"
    mock_go_test
    
    # Source the script and call main function with package and pattern
    run bash -c "source $SCRIPT_PATH && main pkg/services/auth TestAuth"
    
    assert_success
    assert [ -f "${TEST_TEMP_DIR}/go_test_called" ]
    assert_file_contains "${TEST_TEMP_DIR}/go_test_called" "MOCK_GO_TEST_ARGS=-v -tags=integration ./pkg/services/auth/... -run TestAuth"
}
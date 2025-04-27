#!/usr/bin/env bats
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2025 Scott Friedman and Project Contributors

# Load the BATS test helper libraries
load "bats/bats-support/load.bash"
load "bats/bats-assert/load.bash"
load "bats/bats-file/load.bash"

# Path to the scripts
CHECK_SCRIPT_PATH="../scripts/check-license-headers.sh"
UPDATE_SCRIPT_PATH="../scripts/update-license-headers.sh"

# Setup function runs before each test
setup() {
    # Create a temporary directory for tests
    TEST_TEMP_DIR="$(mktemp -d)"
    
    # Export as environment variable for use in tests
    export TEST_TEMP_DIR
    
    # Create sample files with and without license headers
    create_sample_files
}

# Teardown function runs after each test
teardown() {
    # Remove the temporary directory
    rm -rf "$TEST_TEMP_DIR"
}

# Helper function to create sample files
create_sample_files() {
    # Go file with correct header
    mkdir -p "${TEST_TEMP_DIR}/pkg/core"
    cat > "${TEST_TEMP_DIR}/pkg/core/sample.go" << EOF
// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package core

func SampleFunction() string {
    return "Hello, World!"
}
EOF

    # Go file without header
    cat > "${TEST_TEMP_DIR}/pkg/core/missing_header.go" << EOF
package core

func AnotherFunction() string {
    return "Missing header"
}
EOF

    # Shell script with incorrect header
    mkdir -p "${TEST_TEMP_DIR}/scripts"
    cat > "${TEST_TEMP_DIR}/scripts/test.sh" << EOF
#!/bin/bash
# Copyright (c) 2025 Wrong Header
echo "This has a wrong header"
EOF

    # Python file without header
    cat > "${TEST_TEMP_DIR}/scripts/test.py" << EOF
#!/usr/bin/env python3
def main():
    print("No header here")
EOF
}

# Test check-license-headers script basic execution
@test "check-license-headers script runs without errors" {
    cd "$TEST_TEMP_DIR"
    run bash "$CHECK_SCRIPT_PATH" --directory "$TEST_TEMP_DIR" || true
    
    # Even though some files will fail the check, the script should run
    assert_output --partial "Checking license headers"
}

# Test check-license-headers detects missing headers
@test "check-license-headers correctly identifies missing headers" {
    cd "$TEST_TEMP_DIR"
    run bash "$CHECK_SCRIPT_PATH" --directory "$TEST_TEMP_DIR" || true
    
    assert_output --partial "missing_header.go"
    assert_output --partial "test.py"
}

# Test check-license-headers detects incorrect headers
@test "check-license-headers correctly identifies incorrect headers" {
    cd "$TEST_TEMP_DIR"
    run bash "$CHECK_SCRIPT_PATH" --directory "$TEST_TEMP_DIR" || true
    
    assert_output --partial "test.sh"
}

# Test check-license-headers passes correct headers
@test "check-license-headers passes files with correct headers" {
    cd "$TEST_TEMP_DIR"
    run bash "$CHECK_SCRIPT_PATH" --file "${TEST_TEMP_DIR}/pkg/core/sample.go"
    
    assert_success
    assert_output --partial "Header is correct"
}

# Test update-license-headers script adds missing headers
@test "update-license-headers adds missing headers" {
    cd "$TEST_TEMP_DIR"
    
    # Verify header is missing
    assert_not_file_contains "${TEST_TEMP_DIR}/pkg/core/missing_header.go" "SPDX-License-Identifier"
    
    # Run update script
    run bash "$UPDATE_SCRIPT_PATH" --directory "$TEST_TEMP_DIR"
    assert_success
    
    # Verify header was added
    assert_file_contains "${TEST_TEMP_DIR}/pkg/core/missing_header.go" "SPDX-License-Identifier: Apache-2.0"
    assert_file_contains "${TEST_TEMP_DIR}/pkg/core/missing_header.go" "Copyright (c) 2025 Scott Friedman and Project Contributors"
}

# Test update-license-headers corrects incorrect headers
@test "update-license-headers corrects incorrect headers" {
    cd "$TEST_TEMP_DIR"
    
    # Verify wrong header exists
    assert_file_contains "${TEST_TEMP_DIR}/scripts/test.sh" "Copyright (c) 2025 Wrong Header"
    
    # Run update script
    run bash "$UPDATE_SCRIPT_PATH" --directory "$TEST_TEMP_DIR"
    assert_success
    
    # Verify header was corrected
    assert_file_contains "${TEST_TEMP_DIR}/scripts/test.sh" "SPDX-License-Identifier: Apache-2.0"
    assert_file_contains "${TEST_TEMP_DIR}/scripts/test.sh" "Copyright (c) 2025 Scott Friedman and Project Contributors"
    assert_not_file_contains "${TEST_TEMP_DIR}/scripts/test.sh" "Wrong Header"
}
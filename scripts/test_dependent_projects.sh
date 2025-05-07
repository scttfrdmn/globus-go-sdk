#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2025 Scott Friedman and Project Contributors

# This script tests that dependent projects can successfully build with the
# current version of the Globus Go SDK

set -e

# Read the current SDK version
SDK_VERSION=$(grep "const Version" pkg/core/version.go | awk -F'"' '{print $2}')
echo "Testing SDK version: ${SDK_VERSION}"

# Create a temporary directory for testing
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

echo "Created temporary directory: ${TEMP_DIR}"

# Test building the Globus CLI with this SDK version
test_globus_cli() {
    echo "Testing build of globus-go-cli..."
    
    # Clone the repository if CLI_REPO_PATH isn't set
    if [ -z "$CLI_REPO_PATH" ]; then
        CLI_REPO_PATH="${TEMP_DIR}/globus-go-cli"
        git clone https://github.com/scttfrdmn/globus-go-cli.git "$CLI_REPO_PATH"
    fi
    
    # Save current directory
    CURRENT_DIR=$(pwd)
    
    # Go to CLI directory
    cd "$CLI_REPO_PATH"
    
    # Update go.mod to use the current SDK version
    go get github.com/scttfrdmn/globus-go-sdk@v${SDK_VERSION}
    
    # Try to build
    if go build; then
        echo "✅ globus-go-cli builds successfully with SDK v${SDK_VERSION}"
    else
        echo "❌ globus-go-cli fails to build with SDK v${SDK_VERSION}"
        exit 1
    fi
    
    # Return to original directory
    cd "$CURRENT_DIR"
}

# Function to test a minimal example project
test_minimal_example() {
    echo "Testing minimal example project..."
    
    # Create a test directory
    EXAMPLE_DIR="${TEMP_DIR}/minimal-example"
    mkdir -p "$EXAMPLE_DIR"
    
    # Create a simple Go module that uses the SDK
    cd "$EXAMPLE_DIR"
    
    # Initialize module
    go mod init example.com/test
    
    # Add dependency on SDK
    go get github.com/scttfrdmn/globus-go-sdk@v${SDK_VERSION}
    
    # Create a simple Go file that uses each major component of the SDK
    cat > main.go << 'EOF'
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/scttfrdmn/globus-go-sdk/pkg/core/config"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
    "github.com/scttfrdmn/globus-go-sdk/pkg/services/transfer"
)

func main() {
    // Create a configuration
    cfg := config.DefaultConfig()
    cfg.WithAPIVersionCheck(true)
    
    // Create an auth client
    authClient, err := auth.NewClient(
        auth.WithClientID("test-client"),
    )
    if err != nil {
        log.Fatalf("Failed to create auth client: %v", err)
    }
    
    // Create a transfer client
    transferClient, err := transfer.NewClient(
        transfer.WithHTTPDebugging(true),
    )
    if err != nil {
        log.Fatalf("Failed to create transfer client: %v", err)
    }
    
    // Use DefaultMemoryOptimizedOptions to verify SyncChecksum exists
    options := transfer.DefaultMemoryOptimizedOptions()
    
    fmt.Printf("SDK Version: %s\n", authClient.Client.UserAgent)
    fmt.Printf("Transfer options sync level: %d\n", options.SyncLevel)
    
    // Use API functions that were problematic in the past
    authURL := authClient.GetAuthorizationURL("state")
    fmt.Printf("Auth URL: %s\n", authURL)
    
    // Get endpoints (won't actually make a request)
    _, err = transferClient.ListEndpoints(context.Background(), nil)
    if err != nil {
        fmt.Printf("Expected error: %v\n", err)
    }
}
EOF
    
    # Try to build (should succeed even if runtime would fail due to lack of credentials)
    if go build; then
        echo "✅ Minimal example builds successfully with SDK v${SDK_VERSION}"
    else
        echo "❌ Minimal example fails to build with SDK v${SDK_VERSION}"
        exit 1
    fi
}

# Run the tests
test_minimal_example

# Only run CLI test if CLI_REPO_PATH is set or --with-cli flag is provided
if [ -n "$CLI_REPO_PATH" ] || [ "$1" == "--with-cli" ]; then
    test_globus_cli
fi

echo "All dependent project build tests passed!"
exit 0
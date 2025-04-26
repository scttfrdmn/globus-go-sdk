#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2025 Scott Friedman and Project Contributors

# Script to run integration tests for the Globus Go SDK
# This script checks for required environment variables and runs the tests

set -e  # Exit on error

# Function to check if required environment variables are set
check_env_vars() {
  local missing=0
  
  echo "Checking for required environment variables..."
  
  if [ -z "$GLOBUS_TEST_CLIENT_ID" ]; then
    echo "❌ GLOBUS_TEST_CLIENT_ID is not set"
    missing=1
  else
    echo "✅ GLOBUS_TEST_CLIENT_ID is set"
  fi
  
  if [ -z "$GLOBUS_TEST_CLIENT_SECRET" ]; then
    echo "❌ GLOBUS_TEST_CLIENT_SECRET is not set"
    missing=1
  else
    echo "✅ GLOBUS_TEST_CLIENT_SECRET is set"
  fi
  
  # Optional but recommended for transfer tests
  if [ -z "$GLOBUS_TEST_SOURCE_ENDPOINT_ID" ]; then
    echo "⚠️  GLOBUS_TEST_SOURCE_ENDPOINT_ID is not set (transfer tests may be limited)"
  else
    echo "✅ GLOBUS_TEST_SOURCE_ENDPOINT_ID is set"
  fi
  
  if [ -z "$GLOBUS_TEST_DEST_ENDPOINT_ID" ]; then
    echo "⚠️  GLOBUS_TEST_DEST_ENDPOINT_ID is not set (transfer tests may be limited)"
  else
    echo "✅ GLOBUS_TEST_DEST_ENDPOINT_ID is set"
  fi
  
  # Optional for group tests
  if [ -z "$GLOBUS_TEST_GROUP_ID" ]; then
    echo "⚠️  GLOBUS_TEST_GROUP_ID is not set (some group tests may be skipped)"
  else
    echo "✅ GLOBUS_TEST_GROUP_ID is set"
  fi
  
  if [ $missing -eq 1 ]; then
    echo "❌ Some required environment variables are missing."
    echo "Please set them before running integration tests."
    echo "Example:"
    echo "export GLOBUS_TEST_CLIENT_ID=\"your-client-id\""
    echo "export GLOBUS_TEST_CLIENT_SECRET=\"your-client-secret\""
    exit 1
  fi
}

# Function to run tests with a specific pattern
run_tests() {
  local package=$1
  local pattern=$2
  
  if [ -z "$pattern" ]; then
    echo "Running all tests for $package"
    go test -v -tags=integration $package
  else
    echo "Running tests matching $pattern for $package"
    go test -v -tags=integration $package -run $pattern
  fi
}

# Main function
main() {
  # Check for required environment variables
  check_env_vars
  
  echo ""
  echo "Running integration tests..."
  
  if [ $# -eq 0 ]; then
    # Run all integration tests
    echo "Running integration tests for all packages"
    go test -v -tags=integration ./...
  elif [ $# -eq 1 ]; then
    # Run tests for a specific package
    run_tests "./$1/..." ""
  elif [ $# -eq 2 ]; then
    # Run tests for a specific package with a pattern
    run_tests "./$1/..." "$2"
  else
    echo "Usage: $0 [package] [pattern]"
    echo "Examples:"
    echo "  $0                               # Run all integration tests"
    echo "  $0 pkg/services/auth             # Run auth integration tests"
    echo "  $0 pkg/services/transfer Transfer # Run transfer tests with 'Transfer' in the name"
    exit 1
  fi
}

# Run the script
main "$@"
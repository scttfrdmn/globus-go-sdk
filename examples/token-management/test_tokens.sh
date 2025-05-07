#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# This script tests the tokens package with real credentials from .env.test

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SDK_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

echo "=== Running Tokens Package Tests ==="
echo "SDK root: ${SDK_ROOT}"

# Check for .env.test file
if [ ! -f "${SDK_ROOT}/.env.test" ]; then
    echo "Error: .env.test file not found in SDK root directory."
    echo "Please create an .env.test file with your Globus credentials."
    exit 1
fi

# Check for required variables in .env.test
source "${SDK_ROOT}/.env.test"

# Support both prefixed and non-prefixed variables
if [ -n "${GLOBUS_TEST_CLIENT_ID}" ]; then
    export GLOBUS_CLIENT_ID="${GLOBUS_TEST_CLIENT_ID}"
fi
if [ -n "${GLOBUS_TEST_CLIENT_SECRET}" ]; then
    export GLOBUS_CLIENT_SECRET="${GLOBUS_TEST_CLIENT_SECRET}"
fi

if [ -z "${GLOBUS_CLIENT_ID}" ] || [ -z "${GLOBUS_CLIENT_SECRET}" ]; then
    echo "Error: GLOBUS_CLIENT_ID and/or GLOBUS_CLIENT_SECRET missing from .env.test file."
    echo "Make sure you have either GLOBUS_CLIENT_ID or GLOBUS_TEST_CLIENT_ID set."
    exit 1
fi

echo "Found credentials in .env.test"

# Run the Go test program
cd "${SCRIPT_DIR}"
echo "Running token test with credentials..."
go run test_with_credentials.go

# Run the standard example
echo -e "\nRunning standard token management example..."
go run main.go mock.go

echo -e "\n=== All Tests Passed ==="
echo "The tokens package is working correctly!"
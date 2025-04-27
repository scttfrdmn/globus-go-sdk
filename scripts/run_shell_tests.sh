#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2025 Scott Friedman and Project Contributors

# Script to run BATS tests for shell scripts

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Check if BATS is installed
BATS_PATH="$(pwd)/tests/bats/bats-core/bin/bats"
if [ ! -f "$BATS_PATH" ]; then
    echo -e "${YELLOW}BATS not found. Installing...${NC}"
    ./scripts/install_bats.sh
    
    # Check if installation was successful
    if [ ! -f "$BATS_PATH" ]; then
        echo -e "${RED}Failed to install BATS. Please run ./scripts/install_bats.sh manually.${NC}"
        exit 1
    fi
fi

# Find all BATS test files
echo -e "\n${YELLOW}Looking for BATS test files...${NC}"
TEST_FILES=$(find ./tests -name "*.bats" | sort)

if [ -z "$TEST_FILES" ]; then
    echo -e "${RED}No BATS test files found.${NC}"
    exit 0
fi

echo -e "Found $(echo "$TEST_FILES" | wc -l | tr -d '[:space:]') test files:"
echo "$TEST_FILES" | sed 's/^/  - /'

# Check if specific files were provided
if [ $# -gt 0 ]; then
    TEST_FILES="$@"
    echo -e "${YELLOW}Running only the specified test files: ${TEST_FILES}${NC}"
fi

# Run the tests
echo -e "\n${YELLOW}Running BATS tests...${NC}"
"$BATS_PATH" --tap $TEST_FILES

# Exit with the status of the BATS command
exit $?